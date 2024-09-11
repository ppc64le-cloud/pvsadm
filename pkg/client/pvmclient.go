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

	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	"k8s.io/utils/ptr"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/cloudconnection"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/datacenter"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/dhcp"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/events"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/image"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/instance"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/job"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/key"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/network"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/storagetier"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/volume"
)

type PVMClient struct {
	InstanceName string
	InstanceID   string
	Region       string
	Zone         string

	PISession *ibmpisession.IBMPISession

	CloudConnectionClient *cloudconnection.Client
	DatacenterClient      *datacenter.Client
	DHCPClient            *dhcp.Client
	EventsClient          *events.Client
	ImgClient             *image.Client
	InstanceClient        *instance.Client
	JobClient             *job.Client
	KeyClient             *key.Client
	NetworkClient         *network.Client
	StorageTierClient     *storagetier.Client
	VolumeClient          *volume.Client
}

func NewPVMClient(c *Client, instanceID, instanceName string, ep map[string]string) (*PVMClient, error) {
	listServiceInstanceOptions := &resourcecontrollerv2.ListResourceInstancesOptions{
		// TODO: possibility of workspaces to either be of type service_instance or composite_instance.
		// Type: ptr.To(serviceInstance),
		Name: ptr.To(instanceName),
		GUID: ptr.To(instanceID),
	}

	workspaces, _, err := c.ResourceControllerClient.ListResourceInstances(listServiceInstanceOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list the resource instances: %v", err)
	}
	instanceIDorName := instanceID
	if instanceIDorName == "" {
		instanceIDorName = instanceName
	}
	if len(workspaces.Resources) == 0 {
		return nil, fmt.Errorf("no resource instances are available to be listed for ID/Name: %s", instanceIDorName)
	}
	pvmclient := &PVMClient{
		InstanceName: *workspaces.Resources[0].Name,
		InstanceID:   *workspaces.Resources[0].GUID,
		Zone:         *workspaces.Resources[0].RegionID,
	}

	pvmclientOptions := ibmpisession.IBMPIOptions{
		Authenticator: &core.IamAuthenticator{ApiKey: pkg.Options.APIKey},
		Debug:         pkg.Options.Debug,
		UserAccount:   c.User.Account,
		URL:           ep[PIEndpoint],
		Zone:          pvmclient.Zone,
	}

	pvmclient.PISession, err = ibmpisession.NewIBMPISession(&pvmclientOptions)
	if err != nil {
		return nil, err
	}

	pvmclient.DatacenterClient = datacenter.NewClient(pvmclient.PISession, instanceID)
	pvmclient.DHCPClient = dhcp.NewClient(pvmclient.PISession, instanceID)
	pvmclient.EventsClient = events.NewClient(pvmclient.PISession, instanceID)
	pvmclient.ImgClient = image.NewClient(pvmclient.PISession, instanceID)
	pvmclient.InstanceClient = instance.NewClient(pvmclient.PISession, instanceID)
	pvmclient.JobClient = job.NewClient(pvmclient.PISession, instanceID)
	pvmclient.KeyClient = key.NewClient(pvmclient.PISession, instanceID)
	pvmclient.NetworkClient = network.NewClient(pvmclient.PISession, instanceID)
	pvmclient.StorageTierClient = storagetier.NewClient(pvmclient.PISession, pvmclient.InstanceID)
	pvmclient.VolumeClient = volume.NewClient(pvmclient.PISession, instanceID)
	return pvmclient, nil
}

func NewGenericPVMClient(c *Client, instanceID string, session *ibmpisession.IBMPISession) (*PVMClient, error) {
	getServiceInstanceOptions := &resourcecontrollerv2.GetResourceInstanceOptions{
		// Possibility of workspaces to either be of type service_instance or composite_instance for Type, hence filter by ID.
		ID: ptr.To(instanceID),
	}

	workspace, _, err := c.ResourceControllerClient.GetResourceInstance(getServiceInstanceOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list the resource instances: %v", err)
	}

	pvmclient := &PVMClient{InstanceID: instanceID, InstanceName: *workspace.Name, Zone: *workspace.RegionID, PISession: session}
	pvmclient.CloudConnectionClient = cloudconnection.NewClient(pvmclient.PISession, instanceID)
	return pvmclient, nil
}
