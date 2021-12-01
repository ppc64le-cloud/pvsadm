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

package sshkeys

import (
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

type Client struct {
	client     *instance.IBMPIKeyClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	c := &Client{
		instanceID: powerinstanceid,
	}
	c.client = instance.NewIBMPIKeyClient(sess, powerinstanceid)
	return c
}

func (c *Client) Get(id string) (*models.SSHKey, error) {
	return c.client.Get(id, c.instanceID)
}

//func (c *Client) GetAll() (*models.SSHKeys, error) {
//	return c.client.GetAll(c.instanceID)
//}

func (c *Client) Create(name string, sshkey string) (*models.SSHKey, *models.SSHKey, error) {
	return c.client.Create(name, sshkey, c.instanceID)
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id, c.instanceID)
}
