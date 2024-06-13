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

package dhcp

import (
	"context"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

type Client struct {
	session    *ibmpisession.IBMPISession
	client     *instance.IBMPIDhcpClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		session:    sess,
		instanceID: powerinstanceid,
		client:     instance.NewIBMPIDhcpClient(context.Background(), sess, powerinstanceid),
	}
}

func (c *Client) Get(id string) (*models.DHCPServerDetail, error) {
	return c.client.Get(id)
}

func (c *Client) GetAll() (models.DHCPServers, error) {
	return c.client.GetAll()
}

func (c *Client) Create(body *models.DHCPServerCreate) (*models.DHCPServer, error) {
	return c.client.Create(body)
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id)
}
