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
	gohttp "net/http"
	"strings"

	"github.com/IBM-Cloud/bluemix-go"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/catalog"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/controller"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/management"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev2/controllerv2"
	"github.com/IBM-Cloud/bluemix-go/authentication"
	"github.com/IBM-Cloud/bluemix-go/http"
	"github.com/IBM-Cloud/bluemix-go/models"
	"github.com/IBM-Cloud/bluemix-go/rest"
	bxsession "github.com/IBM-Cloud/bluemix-go/session"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"github.com/IBM/go-sdk-core/v5/core"

	//"golang.org/x/oauth2/jwt"
	"github.com/golang-jwt/jwt"
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

type Client struct {
	*bxsession.Session
	User                    *User
	ResourceClientV2        controllerv2.ResourceServiceInstanceRepository
	ResourceClientV1        controller.ResourceServiceInstanceRepository
	ResourceServiceKey      controller.ResourceServiceKeyRepository
	ResouceControllerClient *resourcecontrollerv2.ResourceControllerV2
	ResourceControllerOpts  *resourcecontrollerv2.ResourceControllerV2Options
	ResCatalogAPI           catalog.ResourceCatalogRepository
	ResGroupAPI             management.ResourceGroupRepository
}

func authenticateAPIKey(sess *bxsession.Session) error {
	config := sess.Config
	tokenRefresher, err := authentication.NewIAMAuthRepository(config, &rest.Client{
		DefaultHeader: gohttp.Header{
			"User-Agent": []string{http.UserAgent()},
		},
	})
	if err != nil {
		return err
	}
	return tokenRefresher.AuthenticateAPIKey(config.BluemixAPIKey)
}

type User struct {
	ID         string
	Email      string
	Account    string
	cloudName  string `default:"bluemix"`
	cloudType  string `default:"public"`
	generation int    `default:"2"`
}

func fetchUserDetails(sess *bxsession.Session, generation int) (*User, error) {
	config := sess.Config
	user := User{}
	var bluemixToken string

	if strings.HasPrefix(config.IAMAccessToken, "Bearer") {
		bluemixToken = config.IAMAccessToken[7:len(config.IAMAccessToken)]
	} else {
		bluemixToken = config.IAMAccessToken
	}

	token, err := jwt.Parse(bluemixToken, func(token *jwt.Token) (interface{}, error) {
		return "", nil
	})
	if err != nil && !strings.Contains(err.Error(), "key is of invalid type") {
		return &user, err
	}

	claims := token.Claims.(jwt.MapClaims)
	if email, ok := claims["email"]; ok {
		user.Email = email.(string)
	}
	user.ID = claims["id"].(string)
	user.Account = claims["account"].(map[string]interface{})["bss"].(string)
	iss := claims["iss"].(string)
	if strings.Contains(iss, "https://iam.cloud.ibm.com") {
		user.cloudName = "bluemix"
	} else {
		user.cloudName = "staging"
	}
	user.cloudType = "public"

	user.generation = generation
	return &user, nil
}

func NewClient(apikey, ep string, debug bool) (*Client, error) {
	c := &Client{}

	bxSess, err := bxsession.New(&bluemix.Config{
		BluemixAPIKey:         apikey,
		TokenProviderEndpoint: &ep,
		Debug:                 debug,
	})
	if err != nil {
		return nil, err
	}

	c.Session = bxSess

	err = authenticateAPIKey(bxSess)
	if err != nil {
		return nil, err
	}

	c.User, err = fetchUserDetails(bxSess, 2)
	if err != nil {
		return nil, err
	}

	ctrlv2, err := controllerv2.New(bxSess)
	if err != nil {
		return nil, err
	}

	ctrlv1, err := controller.New(bxSess)
	if err != nil {
		return nil, err
	}

	catalogClient, err := catalog.New(bxSess)
	if err != nil {
		return nil, err
	}

	managementClient, err := management.New(bxSess)
	if err != nil {
		return nil, err
	}

	c.ResourceClientV2 = ctrlv2.ResourceServiceInstanceV2()
	c.ResourceClientV1 = ctrlv1.ResourceServiceInstance()
	c.ResourceServiceKey = ctrlv1.ResourceServiceKey()
	c.ResCatalogAPI = catalogClient.ResourceCatalog()
	c.ResGroupAPI = managementClient.ResourceGroup()
	return c, nil
}

// Func ListServiceInstances, list all available instances of particular servicetype
func (c *Client) ListServiceInstances(serviceType string) (map[string]string, error) {
	svcs, err := c.ResourceClientV2.ListInstances(controllerv2.ServiceInstanceQuery{
		Type: "service_instance",
	})

	if err != nil {
		return nil, err
	}

	instances := make(map[string]string)

	for _, svc := range svcs {
		if svc.Crn.ServiceName == serviceType {
			instances[svc.Name] = svc.Guid
		}
	}
	return instances, nil
}

// func ListWorkspaceInstances is used to retrieve serviceInstances along with their regions.
func (c *Client) ListWorkspaceInstances() (*resourcecontrollerv2.ResourceInstancesList, error) {
	auth, err := core.GetAuthenticatorFromEnvironment(serviceIBMCloud)

	if err != nil {
		return nil, err
	}
	c.ResouceControllerClient, err = resourcecontrollerv2.NewResourceControllerV2(&resourcecontrollerv2.ResourceControllerV2Options{Authenticator: auth})
	if err != nil {
		return nil, err
	}

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

func (c *Client) CreateServiceInstance(instanceName, serviceName, servicePlan, resourceGrp, region string) (string, error) {
	//Check Service using ServiceName and returns []models.Service
	service, err := c.ResCatalogAPI.FindByName(serviceName, true)
	if err != nil {
		return "", err
	}

	//GetServicePlanID takes models.Service as the input and returns serviceplanid as the output
	servicePlanID, err := c.ResCatalogAPI.GetServicePlanID(service[0], servicePlan)
	if err != nil {
		return "", err
	}

	if servicePlanID == "" {
		_, err := c.ResCatalogAPI.GetServicePlan(servicePlan)
		if err != nil {
			return "", err
		}
		servicePlanID = servicePlan
	}

	deployments, err := c.ResCatalogAPI.ListDeployments(servicePlanID)
	if err != nil {
		return "", err
	}

	if len(deployments) == 0 {
		klog.Infof("No deployment found for service plan : %s", servicePlan)
		return "", err
	}

	supportedDeployments := []models.ServiceDeployment{}
	supportedLocations := make(map[string]bool)
	for _, d := range deployments {
		if d.Metadata.RCCompatible {
			deploymentLocation := d.Metadata.Deployment.Location
			supportedLocations[deploymentLocation] = true
			if deploymentLocation == region {
				supportedDeployments = append(supportedDeployments, d)
			}
		}
	}

	if len(supportedDeployments) == 0 {
		locationList := make([]string, 0, len(supportedLocations))
		for l := range supportedLocations {
			locationList = append(locationList, l)
		}
		return "", fmt.Errorf("no deployment found for service plan %s at location %s.\nValid location(s) are: %q.\nUse service instance example if the service is a Cloud Foundry service",
			servicePlan, region, locationList)
	}

	//FindByName returns []models.ResourceGroup
	resGrp, err := c.ResGroupAPI.FindByName(nil, resourceGrp)
	if err != nil {
		return "", err
	}

	klog.Infof("Resource group: %s and ID: %s", resGrp[0].Name, resGrp[0].ID)

	var serviceInstancePayload = controller.CreateServiceInstanceRequest{
		Name:            instanceName,
		ServicePlanID:   servicePlanID,
		ResourceGroupID: resGrp[0].ID,
		TargetCrn:       supportedDeployments[0].CatalogCRN,
	}

	serviceInstance, err := c.ResourceClientV1.CreateInstance(serviceInstancePayload)
	if err != nil {
		return "", err
	}

	klog.Infof("Resource service instance details :%v", serviceInstance)
	klog.Infof("Resource service instanceID :%v", serviceInstance.Crn.ServiceInstance)

	return serviceInstance.Crn.ServiceInstance, nil
}

// DeleteSericeInstance deletes service instances on the IBM Cloud, takes instanceID as input
func (c *Client) DeleteServiceInstance(instanceID string, recursive bool) error {
	err := c.ResourceClientV1.DeleteInstance(instanceID, recursive)
	if err != nil {
		klog.Errorf("failed to delete the instance with id %s because of the error %s", instanceID, err)
		return err
	}
	return nil
}

func (c *Client) GetResourceKeys(instanceID string) ([]models.ServiceKey, error) {
	keys, err := c.ResourceServiceKey.GetKeys("")
	if err != nil {
		return nil, err
	}
	var instance_keys []models.ServiceKey
	for _, key := range keys {
		if key.Crn.ServiceInstance == instanceID && key.State == "active" {
			instance_keys = append(instance_keys, key)
		}
	}
	return instance_keys, nil
}
