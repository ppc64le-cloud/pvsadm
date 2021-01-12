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
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/catalog"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/controller"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/management"
	"github.com/IBM-Cloud/bluemix-go/models"
	"github.com/IBM-Cloud/bluemix-go/session"
	"k8s.io/klog/v2"
)

//Func CreateServiceInstance will create a service instance IBM cloud service on IBM Cloud using resourcegroup api's.
//It will accept bluemix client, service type as input and return instance details.
func CreateServiceInstance(sess *session.Session, instanceName, serviceName, servicePlan, resourceGrp, region string) (string, error) {
	catalogClient, err := catalog.New(sess)

	if err != nil {
		return "", err
	}
	resCatalogAPI := catalogClient.ResourceCatalog()

	service, err := resCatalogAPI.FindByName(serviceName, true)
	if err != nil {
		return "", err
	}

	servicePlanID, err := resCatalogAPI.GetServicePlanID(service[0], servicePlan)
	if err != nil {
		return "", err
	}
	if servicePlanID == "" {
		_, err := resCatalogAPI.GetServicePlan(servicePlan)
		if err != nil {
			return "", err
		}
		servicePlanID = servicePlan
	}
	deployments, err := resCatalogAPI.ListDeployments(servicePlanID)
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
		return "", fmt.Errorf("No deployment found for service plan %s at location %s.\nValid location(s) are: %q.\nUse service instance example if the service is a Cloud Foundry service.",
			servicePlan, region, locationList)
	}

	managementClient, err := management.New(sess)
	if err != nil {
		return "", err
	}

	var resourceGroup models.ResourceGroup
	resGrpAPI := managementClient.ResourceGroup()

	if resourceGrp == "" {
		resourceGroupQuery := management.ResourceGroupQuery{
			Default: true,
		}

		grpList, err := resGrpAPI.List(&resourceGroupQuery)
		if err != nil {
			return "", err
		}
		resourceGroup = grpList[0]

	} else {
		grp, err := resGrpAPI.FindByName(nil, resourceGrp)
		if err != nil {
			return "", err
		}
		resourceGroup = grp[0]
	}
	klog.Infof("Resource group: %s and ID: %s", resourceGroup.Name, resourceGroup.ID)

	controllerClient, err := controller.New(sess)
	if err != nil {
		return "", err
	}

	resServiceInstanceAPI := controllerClient.ResourceServiceInstance()

	var serviceInstancePayload = controller.CreateServiceInstanceRequest{
		Name:            instanceName,
		ServicePlanID:   servicePlanID,
		ResourceGroupID: resourceGroup.ID,
		TargetCrn:       supportedDeployments[0].CatalogCRN,
	}

	serviceInstance, err := resServiceInstanceAPI.CreateInstance(serviceInstancePayload)
	if err != nil {
		return "", err
	}

	klog.Infof("Resource service Instance Details :%v\n", serviceInstance)

	return serviceInstance.Crn.ServiceInstance, nil
}
