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
	"fmt"

	"github.com/IBM/go-sdk-core/core"
	"github.com/IBM/platform-services-go-sdk/iamidentityv1"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	"github.com/IBM/platform-services-go-sdk/resourcemanagerv2"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

const (
	serviceInstance = "service_instance"
	serviceIBMCloud = "IBMCLOUD"
	// PowerVSResourceID is Power VS power-iaas service id, can be retrieved using ibmcloud cli
	// ibmcloud catalog service power-iaas.
	PowerVSResourceID = "abd259f0-9990-11e8-acc8-b9f54a8f1661"

	// PowerVSResourcePlanID is Power VS power-iaas plan id, can be retrieved using ibmcloud cli
	// ibmcloud catalog service power-iaas.
	PowerVSResourcePlanID = "f165dd34-3a40-423b-9d95-e90a23f724dd"
)

var CosResourcePlanIDs = map[string]string{
	"onerate":  "1e4e33e4-cfa6-4f12-9016-be594a6d5f87",
	"lite":     "2fdf0c08-2d32-4f46-84b5-32e0c92fffd8",
	"standard": "744bfc56-d12c-4866-88d5-dac9139e0e5d",
}

type Client struct {
	User                    *User
	ResouceControllerClient *resourcecontrollerv2.ResourceControllerV2
	ResourceManagerClient   *resourcemanagerv2.ResourceManagerV2
	ResourceControllerOpts  *resourcecontrollerv2.ResourceControllerV2Options
}

type User struct {
	ID      string
	Email   string
	Account string
}

func NewClient(apikey string, ep map[string]string, debug bool) (*Client, error) {
	c := &Client{}

	auth, err := core.GetAuthenticatorFromEnvironment(serviceIBMCloud)
	if err != nil {
		return nil, err
	}

	iamv1, err := iamidentityv1.NewIamIdentityV1(&iamidentityv1.IamIdentityV1Options{
		Authenticator: auth,
		URL:           ep["TPEndpoint"],
	})
	if err != nil {
		return nil, err
	}
	apiKeyDetailsOpt := iamidentityv1.GetAPIKeysDetailsOptions{IamAPIKey: ptr.To(apikey)}
	apiKey, _, err := iamv1.GetAPIKeysDetails(&apiKeyDetailsOpt)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, fmt.Errorf("could not retrieve account id")
	}
	c.User = &User{
		ID:      *apiKey.ID,
		Account: *apiKey.AccountID,
	}
	resourceControllerOptions := &resourcecontrollerv2.ResourceControllerV2Options{
		URL:           ep["RCEndpoint"],
		Authenticator: auth}
	c.ResouceControllerClient, err = resourcecontrollerv2.NewResourceControllerV2(resourceControllerOptions)
	if err != nil {
		return nil, err
	}
	c.ResourceManagerClient, err = resourcemanagerv2.NewResourceManagerV2(&resourcemanagerv2.ResourceManagerV2Options{Authenticator: auth})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Func ListServiceInstances, list all available instances of particular servicetype
func (c *Client) ListServiceInstances(resourceId string) (map[string]string, error) {
	instances := make(map[string]string)

	listServiceInstanceOptions := &resourcecontrollerv2.ListResourceInstancesOptions{
		ResourceID: ptr.To(resourceId),
	}

	workspaces, _, err := c.ResouceControllerClient.ListResourceInstances(listServiceInstanceOptions)

	if err != nil {
		klog.Errorf("error while listing resource instances: %+v", err)
		return nil, err
	}

	for _, workspace := range workspaces.Resources {
		instances[*workspace.Name] = *workspace.GUID
	}

	return instances, nil
}

// func ListWorkspaceInstances is used to retrieve serviceInstances along with their regions.
func (c *Client) ListWorkspaceInstances() (*resourcecontrollerv2.ResourceInstancesList, error) {

	listServiceInstanceOptions := &resourcecontrollerv2.ListResourceInstancesOptions{
		Type:           ptr.To(serviceInstance),
		ResourceID:     ptr.To(PowerVSResourceID),
		ResourcePlanID: ptr.To(PowerVSResourcePlanID),
	}
	workspaces, _, err := c.ResouceControllerClient.ListResourceInstances(listServiceInstanceOptions)
	if err != nil {
		klog.Errorf("error while listing resource instances: %+v", err)
		return nil, err
	}
	return workspaces, nil

}

func (c *Client) CreateServiceInstance(instanceName, serviceName, resourcePlan, resourceGrp, region string) (*resourcecontrollerv2.ResourceInstance, error) {
	rmv2ListResourceGroupOpt := &resourcemanagerv2.ListResourceGroupsOptions{
		//	AccountID: &c.User.Account,
		Name: &resourceGrp,
	}

	resourceGroupList, _, err := c.ResourceManagerClient.ListResourceGroups(rmv2ListResourceGroupOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to list the reclamations instances: %v", err)
	}

	resourceGroup := resourceGroupList.Resources[0]

	klog.Infof("Resource group: %s and ID: %s", *resourceGroup.Name, *resourceGroup.ID)

	createServiceInstanceOpts := &resourcecontrollerv2.CreateResourceInstanceOptions{
		Name:           ptr.To(instanceName),
		ResourcePlanID: ptr.To(CosResourcePlanIDs[resourcePlan]),
		ResourceGroup:  resourceGroup.ID,
		Target:         ptr.To(region),
	}
	resp, _, err := c.ResouceControllerClient.CreateResourceInstance(createServiceInstanceOpts)
	if err != nil {
		klog.Errorf("An error occured while creating service instance: %v", err)
	}
	if err != nil {
		return nil, err
	}

	klog.Infof("Resource service Instance Details :%+v\n", resp)
	klog.Infof("Resource service InstanceID :%v\n", resp.ID)

	return resp, nil
}

// DeleteSericeInstance deletes service instances on the IBM Cloud, takes instanceID as input
func (c *Client) DeleteServiceInstance(instanceID string, recursive bool) error {
	deleteServiceInstanceOpts := &resourcecontrollerv2.DeleteResourceInstanceOptions{
		ID:        ptr.To(instanceID),
		Recursive: ptr.To(recursive),
	}
	if _, err := c.ResouceControllerClient.DeleteResourceInstance(deleteServiceInstanceOpts); err != nil {
		klog.Errorf("An error occured while deleting service instance: %v", err)
		return err
	}
	return nil
}
