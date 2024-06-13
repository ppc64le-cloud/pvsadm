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

package instance

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/ppc64le-cloud/pvsadm/pkg"
)

type Client struct {
	client     *instance.IBMPIInstanceClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		instanceID: powerinstanceid,
		client:     instance.NewIBMPIInstanceClient(context.Background(), sess, powerinstanceid),
	}
}

func (c *Client) Get(id string) (*models.PVMInstance, error) {
	return c.client.Get(id)
}

func (c *Client) GetAll() (*models.PVMInstances, error) {
	return c.client.GetAll()
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id)
}

func (c *Client) GetAllPurgeable(before, since time.Duration, expr string) ([]*models.PVMInstanceReference, error) {
	instances, err := c.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of instances: %v", err)
	}

	var candidates []*models.PVMInstanceReference
	for _, ins := range instances.PvmInstances {
		if expr != "" {
			if r, _ := regexp.Compile(expr); !r.MatchString(*ins.ServerName) {
				continue
			}
		}
		if !pkg.IsPurgeable(time.Time(ins.CreationDate), before, since) {
			continue
		}
		candidates = append(candidates, ins)
	}
	return candidates, nil
}
