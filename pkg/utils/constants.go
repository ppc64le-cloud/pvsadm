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

package utils

const (
	ServiceTypeCloudObjectStorage = "cloud-object-storage"
)

const (
	// CosResourceID is IBM COS service id, can be retrieved using ibmcloud cli
	// ibmcloud catalog service cloud-object-storage.
	CosResourceID = "dff97f5c-bc5e-4455-b470-411c3edbe49c"

	// PowerVSResourceID is Power VS power-iaas service id, can be retrieved using ibmcloud cli
	// ibmcloud catalog service power-iaas.
	PowerVSResourceID = "abd259f0-9990-11e8-acc8-b9f54a8f1661"

	// PowerVSResourcePlanID is Power VS power-iaas plan id, can be retrieved using ibmcloud cli
	// ibmcloud catalog service power-iaas.
	PowerVSResourcePlanID = "f165dd34-3a40-423b-9d95-e90a23f724dd"
)

var CosResourcePlanIDs = map[string]string{
	"onerate":  "1e4e33e4-cfa6-4f12-9016-be594a6d5f87",
	"lite":     "2fdf0c08-2d32-4f46-84b5-32e0c92fffd8",
	"standard": "744bfc56-d12c-4866-88d5-dac9139e0e5d",
}

const (
	DeletePromptMessage = "Deleting all the above %s and the action is irreversible. Do you really want to continue?"
)
