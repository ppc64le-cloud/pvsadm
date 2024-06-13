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

package volume

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_volumes"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/go-openapi/strfmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
)

type Client struct {
	session    *ibmpisession.IBMPISession
	client     *instance.IBMPIVolumeClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		session:    sess,
		instanceID: powerinstanceid,
		client:     instance.NewIBMPIVolumeClient(context.Background(), sess, powerinstanceid),
	}
}

func (c *Client) Get(id string) (*models.Volume, error) {
	return c.client.Get(id)
}

func (c *Client) DeleteVolume(id string) error {
	return c.client.DeleteVolume(id)
}

func (c *Client) GetAll() (*models.Volumes, error) {
	klog.V(1).Info("Calling the Power Volumes GetAll Method")
	params := p_cloud_volumes.NewPcloudCloudinstancesVolumesGetallParamsWithTimeout(pkg.TIMEOUT).WithCloudInstanceID(c.instanceID)
	resp, err := c.session.Power.PCloudVolumes.PcloudCloudinstancesVolumesGetall(params, c.session.AuthInfo(c.instanceID))
	if err != nil {
		return nil, errors.ToError(err)
	}
	return resp.Payload, nil
}

func (c *Client) getAllPurgeable(field string, before, since time.Duration, expr string) ([]*models.VolumeReference, error) {
	volumes, err := c.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of volumes: %v", err)
	}

	var candidates []*models.VolumeReference
	for _, vol := range volumes.Volumes {
		if expr != "" {
			if r, _ := regexp.Compile(expr); !r.MatchString(*vol.Name) {
				continue
			}
		}
		r := reflect.ValueOf(vol)
		f := reflect.Indirect(r).FieldByName(field)
		fieldValue := f.Interface()
		if !pkg.IsPurgeable(time.Time(*fieldValue.(*strfmt.DateTime)), before, since) {
			continue
		}
		candidates = append(candidates, vol)
	}
	return candidates, nil
}

// Returns all the Purgeable volumes by Last Update Date
func (c *Client) GetAllPurgeableByLastUpdateDate(before, since time.Duration, expr string) ([]*models.VolumeReference, error) {
	return c.getAllPurgeable("LastUpdateDate", before, since, expr)
}
