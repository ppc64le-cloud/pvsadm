// Copyright 2022 IBM Corp
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

package sync

import (
	"errors"
	"os"
	"strconv"
	"testing"

	mocksync "github.com/ppc64le-cloud/pvsadm/cmd/image/sync/mock"
	pkg "github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// Test case constants
const (
	numSources          = 3
	numTargetsPerSource = 3
	numObjects          = 200
)

func TestCalculateChannels(t *testing.T) {
	t.Run("Calculate Channels", func(t *testing.T) {
		// creating mock controller object
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// Create object for mock client
		mockSyncClient := mocksync.NewMockSyncClient(mockCtrl)

		// test case setup
		mockSetBucketLocationConstraint(mockSyncClient, numSources, true, "")
		mockSetSelectedObjects(mockSyncClient, numObjects, numSources)

		// generating spec slice
		spec := mockCreateSpec()

		// generating necessary instance slice
		instanceList := mockCreateInstances(mockSyncClient)

		// test case verification section
		totalChannels, err := calculateChannels(spec, instanceList)
		require.NoError(t, err, "Error calculating channels")
		assert.Equal(t, numObjects*numSources*numTargetsPerSource, totalChannels)
	})
}

func TestGetSpec(t *testing.T) {
	t.Run("Get Specifications", func(t *testing.T) {
		// creating mock controller object
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// generating spec slice
		spec := mockCreateSpec()

		// test case verification section
		file, err := os.CreateTemp("", "spec.*.yaml")
		require.NoError(t, err, "Error creating specfile")
		defer file.Close()

		SpecFileName := file.Name()
		defer os.Remove(SpecFileName)
		klog.V(1).Infof("Specfile :%s", SpecFileName)

		specString, merr := yaml.Marshal(&spec)
		require.NoError(t, merr, "Error in Unmarshalling Specfile")

		_, err = file.WriteString(string(specString))
		require.NoError(t, err, "Error Writing Spec string to file")

		specRes, err := getSpec(SpecFileName)
		require.NoError(t, err, "Error Getting Specification struct")

		for srcIdx, src := range spec {
			assert.Equal(t, src.Source, specRes[srcIdx].Source)
			assert.Equal(t, src.Bucket, specRes[srcIdx].Bucket)
			assert.Equal(t, src.Cos, specRes[srcIdx].Cos)
			assert.Equal(t, src.Region, specRes[srcIdx].Region)
			assert.Equal(t, src.Object, specRes[srcIdx].Object)

			for tgtIdx, tgt := range src.Target {
				assert.Equal(t, tgt.Bucket, specRes[srcIdx].Target[tgtIdx].Bucket)
				assert.Equal(t, tgt.Region, specRes[srcIdx].Target[tgtIdx].Region)
				assert.Equal(t, tgt.StorageClass, specRes[srcIdx].Target[tgtIdx].StorageClass)
			}
		}
	})
}

func TestSync(t *testing.T) {
	tests := []struct {
		name           string
		instanceList   func(mockSyncClient *mocksync.MockSyncClient) []InstanceItem
		spec           func() []pkg.Spec
		mockSyncClient func(mockCtrl *gomock.Controller) *mocksync.MockSyncClient
		setup          func(mockSyncClient *mocksync.MockSyncClient)
		expectedError  string
	}{

		{
			name: "Sync Objects",
			mockSyncClient: func(mockCtrl *gomock.Controller) *mocksync.MockSyncClient {
				mockSyncClient := mocksync.NewMockSyncClient(mockCtrl)
				return mockSyncClient
			},
			instanceList: func(mockSyncClient *mocksync.MockSyncClient) []InstanceItem {
				return mockCreateInstances(mockSyncClient)
			},
			spec: func() []pkg.Spec {
				return mockCreateSpec()
			},
			setup: func(mockSyncClient *mocksync.MockSyncClient) {
				mockSetBucketLocationConstraint(mockSyncClient, numSources, true, "")
				mockSetSelectedObjects(mockSyncClient, numObjects, numSources)
				mockSetSelectedObjects(mockSyncClient, numObjects, numSources)
				mockSetBucketLocationConstraint(mockSyncClient, numSources*numTargetsPerSource, true, "")
				mockSetCopyObjectToBucket(mockSyncClient, numObjects*numSources*numTargetsPerSource, "")

			},
			expectedError: "",
		},

		{
			name: "No Objects Selected",
			mockSyncClient: func(mockCtrl *gomock.Controller) *mocksync.MockSyncClient {
				mockSyncClient := mocksync.NewMockSyncClient(mockCtrl)
				return mockSyncClient
			},
			instanceList: func(mockSyncClient *mocksync.MockSyncClient) []InstanceItem {
				return mockCreateInstances(mockSyncClient)
			},
			spec: func() []pkg.Spec {
				return mockCreateSpec()
			},
			setup: func(mockSyncClient *mocksync.MockSyncClient) {
				mockSetBucketLocationConstraint(mockSyncClient, numSources, true, "")
				mockSetSelectedObjects(mockSyncClient, 0, numSources)
				mockSetSelectedObjects(mockSyncClient, 0, numSources)
				mockSetBucketLocationConstraint(mockSyncClient, numSources*numTargetsPerSource, true, "")
				mockSetCopyObjectToBucket(mockSyncClient, 0, "")

			},
			expectedError: "",
		},

		{
			name: "Bucket Location constraint verification fails for source",
			mockSyncClient: func(mockCtrl *gomock.Controller) *mocksync.MockSyncClient {
				mockSyncClient := mocksync.NewMockSyncClient(mockCtrl)
				return mockSyncClient
			},
			instanceList: func(mockSyncClient *mocksync.MockSyncClient) []InstanceItem {
				return mockCreateInstances(mockSyncClient)
			},
			spec: func() []pkg.Spec {
				return mockCreateSpec()
			},
			setup: func(mockSyncClient *mocksync.MockSyncClient) {
				mockSetBucketLocationConstraint(mockSyncClient, 1, false, "Failed to verify bucket location constraint")
			},
			expectedError: "Failed to verify bucket location constraint",
		},

		{
			name: "Bucket Location constraint verification fails for a target",
			mockSyncClient: func(mockCtrl *gomock.Controller) *mocksync.MockSyncClient {
				mockSyncClient := mocksync.NewMockSyncClient(mockCtrl)
				return mockSyncClient
			},
			instanceList: func(mockSyncClient *mocksync.MockSyncClient) []InstanceItem {
				return mockCreateInstances(mockSyncClient)
			},
			spec: func() []pkg.Spec {
				return mockCreateSpec()
			},
			setup: func(mockSyncClient *mocksync.MockSyncClient) {
				mockSetBucketLocationConstraint(mockSyncClient, numSources, true, "")
				mockSetSelectedObjects(mockSyncClient, numObjects, numSources)
				mockSetSelectedObjects(mockSyncClient, numObjects, 1)
				mockSetBucketLocationConstraint(mockSyncClient, 1, false, "Failed to verify bucket location constriant")
			},
			expectedError: "bucket location constraint verification failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// creating mock controller object
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// Create object for mock client
			mockSyncClient := test.mockSyncClient(mockCtrl)

			// test case setup
			test.setup(mockSyncClient)

			// generating spec slice
			spec := test.spec()

			// generating necessary instance slice
			instanceList := test.instanceList(mockSyncClient)

			// test case verification section
			err := syncObjects(spec, instanceList)
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err, "error syncing objects")
			}
		})
	}
}

func mockCreateInstances(mockSyncClient *mocksync.MockSyncClient) []InstanceItem {
	var instanceList []InstanceItem
	for i := 0; i < numSources; i++ {
		instance := InstanceItem{}
		instance.Source = mockSyncClient

		for j := 0; j < numTargetsPerSource; j++ {
			instance.Target = append(instance.Target, mockSyncClient)
		}
		instanceList = append(instanceList, instance)
	}
	return instanceList
}

func mockCreateSpec() []pkg.Spec {
	klog.V(1).Info("STEP: Generating Spec")
	specSlice := make([]pkg.Spec, 0)
	for i := 0; i < numSources; i++ {
		specSlice = append(specSlice, utils.GenerateSpec(numTargetsPerSource))
	}
	return specSlice
}

func mockSetSelectedObjects(mockSyncClient *mocksync.MockSyncClient, objectsCount int, times int) {
	var res []string
	for i := 0; i < objectsCount; i++ {
		res = append(res, "obj-test"+strconv.Itoa(i)+".iso")
	}

	mockSyncClient.EXPECT().SelectObjects(gomock.Any(), gomock.Any()).Return(
		res, nil,
	).Times(times)
}

func mockSetBucketLocationConstraint(mockSyncClient *mocksync.MockSyncClient, times int, pass bool, err string) {
	if pass {
		mockSyncClient.EXPECT().CheckBucketLocationConstraint(gomock.Any(), gomock.Any()).Return(
			true, nil,
		).Times(times)
	} else {
		mockSyncClient.EXPECT().CheckBucketLocationConstraint(gomock.Any(), gomock.Any()).Return(
			false, errors.New(err),
		).Times(times)
	}
}

func mockSetCopyObjectToBucket(mockSyncClient *mocksync.MockSyncClient, times int, err string) {
	if err == "" {
		mockSyncClient.EXPECT().CopyObjectToBucket(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			nil,
		).Times(times)
	} else {
		mockSyncClient.EXPECT().CopyObjectToBucket(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			errors.New("Copy Objects failed"),
		).Times(times)
	}
}
