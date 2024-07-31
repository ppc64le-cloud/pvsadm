// Copyright 2024 IBM Corp
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

package cloudconnection

import (
	"context"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

type Client struct {
	client     *instance.IBMPICloudConnectionClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		instanceID: powerinstanceid,
		client:     instance.NewIBMPICloudConnectionClient(context.Background(), sess, powerinstanceid),
	}
}

func (c *Client) Get(id string) (*models.CloudConnection, error) {
	return c.client.Get(id)
}

func (c *Client) GetAll() (*models.CloudConnections, error) {
	return c.client.GetAll()
}
