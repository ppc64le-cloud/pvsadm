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

package key

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/ppc64le-cloud/pvsadm/pkg"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

type Client struct {
	session    *ibmpisession.IBMPISession
	client     *instance.IBMPIKeyClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		session:    sess,
		instanceID: powerinstanceid,
		client:     instance.NewIBMPIKeyClient(context.Background(), sess, powerinstanceid),
	}
}

func (c *Client) Get(id string) (*models.SSHKey, error) {
	return c.client.Get(id)
}

func (c *Client) Create(body *models.SSHKey) (*models.SSHKey, error) {
	return c.client.Create(body)
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id)
}

func (c *Client) GetAllPurgeable(before, since time.Duration, expr string) ([]string, error) {
	keys, err := c.client.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of instances: %v", err)
	}

	var keysMatched []string
	r, _ := regexp.Compile(expr)

	for _, key := range keys.SSHKeys {
		if !r.MatchString(*key.Name) {
			continue
		}
		if !pkg.IsPurgeable(time.Time(*key.CreationDate), before, since) {
			continue
		}
		keysMatched = append(keysMatched, *key.Name)
	}
	return keysMatched, nil
}
