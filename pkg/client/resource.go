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

	var resourceGroupID string
	resGrpAPI := managementClient.ResourceGroup()

	if resourceGrp == "" {
		resourceGroupQuery := management.ResourceGroupQuery{
			Default: true,
		}

		grpList, err := resGrpAPI.List(&resourceGroupQuery)
		if err != nil {
			return "", err
		}
		resourceGroupID = grpList[0].ID

	} else {
		grp, err := resGrpAPI.FindByName(nil, resourceGrp)
		if err != nil {
			return "", err
		}
		resourceGroupID = grp[0].ID
	}

	controllerClient, err := controller.New(sess)
	if err != nil {
		return "", err
	}

	resServiceInstanceAPI := controllerClient.ResourceServiceInstance()

	var serviceInstancePayload = controller.CreateServiceInstanceRequest{
		Name:            instanceName,
		ServicePlanID:   servicePlanID,
		ResourceGroupID: resourceGroupID,
		TargetCrn:       supportedDeployments[0].CatalogCRN,
	}

	serviceInstance, err := resServiceInstanceAPI.CreateInstance(serviceInstancePayload)
	if err != nil {
		return "", err
	}

	klog.Infof("Resource service Instance Details :%v\n", serviceInstance)

	return serviceInstance.Crn.ServiceInstance, nil
}
