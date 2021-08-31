package sync

import (
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
)

//go:generate mockgen -source=./sync_client.go -destination=./mock/sync_client_generated.go -package=mock

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
