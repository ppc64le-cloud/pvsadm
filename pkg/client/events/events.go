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
	"time"

	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_events"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/ppc64le-cloud/pvsadm/pkg"
)

type Client struct {
	instanceID string
	client     p_cloud_events.ClientService
	session    *ibmpisession.IBMPISession
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	return &Client{
		session:    sess,
		instanceID: powerinstanceid,
		client:     sess.Power.PCloudEvents,
	}
}

func (c *Client) GetPcloudEventsGetsince(since time.Duration) (*p_cloud_events.PcloudEventsGetqueryOK, error) {
	params := p_cloud_events.NewPcloudEventsGetqueryParamsWithTimeout(pkg.TIMEOUT).WithCloudInstanceID(c.instanceID).WithFromTime(core.StringPtr(time.Now().UTC().Add(-since).Format(time.RFC3339)))
	return c.client.PcloudEventsGetquery(params, c.session.AuthInfo(c.instanceID))
}
