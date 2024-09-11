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

package network

import (
	"context"
	"fmt"
	"regexp"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

type Client struct {
	session    *ibmpisession.IBMPISession
	client     *instance.IBMPINetworkClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		session:    sess,
		instanceID: powerinstanceid,
		client:     instance.NewIBMPINetworkClient(context.Background(), sess, powerinstanceid),
	}
}

func (c *Client) Get(id string) (*models.Network, error) {
	return c.client.Get(id)
}

func (c *Client) GetAllPublic() (*models.Networks, error) {
	return c.client.GetAllPublic()
}

func (c *Client) GetAll() (*models.Networks, error) {
	return c.client.GetAll()
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id)
}

func (c *Client) GetAllPurgeable(expr string) ([]*models.NetworkReference, error) {
	networks, err := c.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of instances: %v", err)
	}

	var candidates []*models.NetworkReference
	for _, network := range networks.Networks {
		if expr != "" {
			if r, _ := regexp.Compile(expr); !r.MatchString(*network.Name) {
				continue
			}
		}
		candidates = append(candidates, network)
	}
	return candidates, nil
}

func (c *Client) CreatePort(id string, params *models.NetworkPortCreate) (*models.NetworkPort, error) {
	return c.client.CreatePort(id, params)
}

func (c *Client) DeletePort(id, portID string) error {
	return c.client.DeletePort(id, portID)
}

func (c *Client) GetPort(id, portID string) (*models.NetworkPort, error) {
	return c.client.GetPort(id, portID)
}

func (c *Client) GetAllPorts(id string) (*models.NetworkPorts, error) {
	return c.client.GetAllPorts(id)
}
