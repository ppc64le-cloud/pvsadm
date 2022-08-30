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
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
)

//go:generate mockgen -source=./sync_client.go -destination=./mock/sync_client_generated.go -package=mock -copyright_file=../../../hack/copyright_file

// sync client interface
type SyncClient interface {
	// S3Client methods
	CopyObjectToBucket(srcBucketName string, destBucketName string, objectName string) error
	CheckBucketLocationConstraint(bucketName string, bucketLocationConstraint string) (bool, error)
	SelectObjects(bucketName string, regex string) ([]string, error)
}

type syncS3Client struct {
	s3 client.S3Client
}

func (c *syncS3Client) CopyObjectToBucket(srcBucketName string, destBucketName string, objectName string) error {
	return c.s3.CopyObjectToBucket(srcBucketName, destBucketName, objectName)
}

func (c *syncS3Client) CheckBucketLocationConstraint(bucketName string, bucketLocationConstraint string) (bool, error) {
	return c.s3.CheckBucketLocationConstraint(bucketName, bucketLocationConstraint)
}

func (c *syncS3Client) SelectObjects(bucketName string, regex string) ([]string, error) {
	return c.s3.SelectObjects(bucketName, regex)
}

func NewS3Client(c *client.Client, instanceName string, region string) (SyncClient, error) {
	s3Cli, err := client.NewS3Client(c, instanceName, region)
	if err != nil {
		return nil, err
	}

	return &syncS3Client{
		s3: *s3Cli,
	}, nil
}
