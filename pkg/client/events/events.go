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

package events

import (
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_events"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"time"
)

type Client struct {
	instanceID string
	client     *p_cloud_events.Client
	session    *ibmpisession.IBMPISession
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	c := &Client{
		session:    sess,
		instanceID: powerinstanceid,
		client:     sess.Power.PCloudEvents,
	}
	return c
}

func (c *Client) GetPcloudEventsGetsince(since time.Duration) (*p_cloud_events.PcloudEventsGetsinceOK, error) {
	params := p_cloud_events.NewPcloudEventsGetsinceParamsWithTimeout(pkg.TIMEOUT).WithCloudInstanceID(c.instanceID).WithTime(time.Now().UTC().Add(-since).Format(time.RFC3339))
	return c.client.PcloudEventsGetsince(params, ibmpisession.NewAuth(c.session, c.instanceID))
}
