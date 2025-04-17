// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	"github.com/IBM/platform-services-go-sdk/resourcemanagerv2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

const (
	serviceInstance = "service_instance"
	serviceIBMCloud = "IBMCLOUD"
)

type Client struct {
	User                     *User
	ResourceControllerClient *resourcecontrollerv2.ResourceControllerV2
	ResourceManagerClient    *resourcemanagerv2.ResourceManagerV2
	ResourceControllerOpts   *resourcecontrollerv2.ResourceControllerV2Options
}

type User struct {
	Account string
}

func NewClient(apikey string, ep map[string]string, debug bool) (*Client, error) {
	c := &Client{}
	auth, err := GetAuthenticator()
	if err != nil {
		return nil, err
	}
	accId, err := GetAccountID(auth)
	if err != nil {
		return nil, err
	}

	c.User = &User{
		Account: accId,
	}

	resourceControllerOptions := &resourcecontrollerv2.ResourceControllerV2Options{
		URL:           ep[RCEndpoint],
		Authenticator: auth,
	}
	c.ResourceControllerClient, err = resourcecontrollerv2.NewResourceControllerV2(resourceControllerOptions)
	if err != nil {
		return nil, err
	}
	c.ResourceManagerClient, err = resourcemanagerv2.NewResourceManagerV2(&resourcemanagerv2.ResourceManagerV2Options{Authenticator: auth})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListServiceInstances list all available instances of particular servicetype
func (c *Client) ListServiceInstances(resourceId string) (map[string]string, error) {
	serviceInstances := make(map[string]string)

	listServiceInstanceOptions := &resourcecontrollerv2.ListResourceInstancesOptions{
		ResourceID: ptr.To(resourceId),
	}

	cloudResource, _, err := c.ResourceControllerClient.ListResourceInstances(listServiceInstanceOptions)

	if err != nil {
		klog.Errorf("error while listing resource instances: %v", err)
		return nil, err
	}
	for _, resource := range cloudResource.Resources {
		serviceInstances[*resource.Name] = *resource.GUID
	}

	return serviceInstances, nil
}

// ListWorkspaceInstances is used to retrieve serviceInstances along with their regions.
func (c *Client) ListWorkspaceInstances() (*resourcecontrollerv2.ResourceInstancesList, error) {

	listServiceInstanceOptions := &resourcecontrollerv2.ListResourceInstancesOptions{
		Type:           ptr.To(serviceInstance),
		ResourceID:     ptr.To(utils.PowerVSResourceID),
		ResourcePlanID: ptr.To(utils.PowerVSResourcePlanID),
	}
	workspaces, _, err := c.ResourceControllerClient.ListResourceInstances(listServiceInstanceOptions)
	if err != nil {
		klog.Errorf("error while listing resource instances: %+v", err)
		return nil, err
	}
	return workspaces, nil

}

func (c *Client) CreateServiceInstance(instanceName, serviceName, resourcePlanID, resourceGrp, region string) (*resourcecontrollerv2.ResourceInstance, error) {
	rmv2ListResourceGroupOpt := &resourcemanagerv2.ListResourceGroupsOptions{
		Name: &resourceGrp,
	}

	resourceGroupList, _, err := c.ResourceManagerClient.ListResourceGroups(rmv2ListResourceGroupOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to list the resource groups: %v", err)
	}
	if len(resourceGroupList.Resources) == 0 {
		return nil, fmt.Errorf("no resource groups were present")
	}

	resourceGroup := resourceGroupList.Resources[0]

	klog.Infof("Resource group: %s and ID: %s", *resourceGroup.Name, *resourceGroup.ID)

	createServiceInstanceOpts := &resourcecontrollerv2.CreateResourceInstanceOptions{
		Name:           ptr.To(instanceName),
		ResourcePlanID: ptr.To(resourcePlanID),
		ResourceGroup:  resourceGroup.ID,
		Target:         ptr.To(region),
	}
	resp, _, err := c.ResourceControllerClient.CreateResourceInstance(createServiceInstanceOpts)
	if err != nil {
		klog.Errorf("An error occured while creating service instance: %v", err)
		return nil, err
	}
	klog.Infof("Created service instance %s of %s-%s in %s", instanceName, serviceName, resourcePlanID, region)
	return resp, nil
}

// DeleteServiceInstance deletes service instances on the IBM Cloud, takes instanceID as input
func (c *Client) DeleteServiceInstance(instanceID string, recursive bool) error {
	deleteServiceInstanceOpts := &resourcecontrollerv2.DeleteResourceInstanceOptions{
		ID:        ptr.To(instanceID),
		Recursive: ptr.To(recursive),
	}
	if _, err := c.ResourceControllerClient.DeleteResourceInstance(deleteServiceInstanceOpts); err != nil {
		klog.Errorf("An error occured while deleting service instance: %v", err)
		return err
	}
	return nil
}

func GetAccountID(auth core.Authenticator) (string, error) {
	// fake request to get a bearer token from the request header
	ctx := context.TODO()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", http.NoBody)
	if err != nil {
		return "", err
	}
	err = auth.Authenticate(req)
	if err != nil {
		return "", err
	}

	bearerToken := req.Header.Get("Authorization")
	if strings.HasPrefix(bearerToken, "Bearer") {
		bearerToken = bearerToken[7:]
	}
	token, err := jwt.Parse(bearerToken, func(_ *jwt.Token) (interface{}, error) {
		return "", nil
	})

	if err != nil && !strings.Contains(err.Error(), "key is of invalid type") {
		return "", err
	}

	return token.Claims.(jwt.MapClaims)["account"].(map[string]interface{})["bss"].(string), nil
}

func GetAuthenticator() (core.Authenticator, error) {
	auth, err := core.GetAuthenticatorFromEnvironment(serviceIBMCloud)
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return nil, fmt.Errorf("authenticator can't be nil, please set proper authentication")
	}
	return auth, nil
}
