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

package client

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev2/controllerv2"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
)

type S3Client struct {
	ApiKey       string
	InstanceName string
	InstanceID   string
	Region       string
	StorageClass string
	SvcEndpoint  string
	S3Session    *s3.S3
}

const (
	AuthEndpoint = "https://iam.cloud.ibm.com/identity/token"
)

//Func NewS3Client accepts apikey, accesskey, secretkey of the bucket and return the s3 client
//to perform different s3 operations like upload, delete etc.,
func NewS3Clientwithkeys(accesskey, secretkey, region string) (s3client *S3Client, err error) {

	s3client = &S3Client{}
	s3client.SvcEndpoint = fmt.Sprintf("https://s3.%s.cloud-object-storage.appdomain.cloud", region)
	s3client.StorageClass = fmt.Sprintf("%s-standard", region)
	conf := aws.NewConfig().
		WithRegion(s3client.StorageClass).
		WithEndpoint(s3client.SvcEndpoint).
		WithCredentials(credentials.NewStaticCredentials(accesskey, secretkey, "")).
		WithS3ForcePathStyle(true)

	// Create client connection
	sess := session.Must(session.NewSession()) // Creating a new session
	s3client.S3Session = s3.New(sess, conf)    // Creating a new client
	return s3client, nil

}

//Func NewS3Client accepts apikey, instanceid of the IBM COS instance and return the s3 client
//to perform different s3 operations like upload, delete etc.,
func NewS3Client(c *Client, instanceName, region string) (s3client *S3Client, err error) {
	s3client = &S3Client{}
	var instanceID string
	svcs, err := c.ResourceClientV2.ListInstances(controllerv2.ServiceInstanceQuery{
		Type: "service_instance",
		Name: instanceName,
	})
	if err != nil {
		return s3client, fmt.Errorf("failed to list the resource instances: %v", err)
	}
	found := false
	for _, svc := range svcs {
		klog.V(4).Infof("Service ID: %s, region_id: %s, Name: %s", svc.Guid, svc.RegionID, svc.Name)
		klog.V(4).Infof("crn: %v", svc.Crn)
		if svc.Name == instanceName {
			instanceID = svc.Guid
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("instance: %s not found", instanceName)
	}
	s3client.InstanceID = instanceID

	if pkg.Options.APIKey == "" {
		s3client.ApiKey = os.Getenv("IBMCLOUD_API_KEY")
	} else {
		s3client.ApiKey = pkg.Options.APIKey
	}
	s3client.SvcEndpoint = fmt.Sprintf("https://s3.%s.cloud-object-storage.appdomain.cloud", region)
	s3client.StorageClass = fmt.Sprintf("%s-standard", region)
	conf := aws.NewConfig().
		WithRegion(s3client.StorageClass).
		WithEndpoint(s3client.SvcEndpoint).
		WithCredentials(ibmiam.NewStaticCredentials(aws.NewConfig(), AuthEndpoint, s3client.ApiKey, s3client.InstanceID)).
		WithS3ForcePathStyle(true)

	// Create client connection
	sess := session.Must(session.NewSession()) // Creating a new session
	s3client.S3Session = s3.New(sess, conf)    // Creating a new client
	return s3client, nil
}

//Func CheckBucketExists will verify for the existence of the bucket in the particular account
func (c *S3Client) CheckBucketExists(bucketName string) (bool, error) {
	result, err := c.S3Session.ListBuckets(nil)
	if err != nil {
		klog.Infof("Unable to list buckets, %v\n", err)
		return false, err
	}

	bucketExists := false
	for _, b := range result.Buckets {
		if aws.StringValue(b.Name) == bucketName {
			bucketExists = true
		}
	}

	if bucketExists {
		return true, nil
	}
	return false, nil
}

// To select objects matching regex from src bucket
func (c *S3Client) SelectObjects(bucketName string, regex string) ([]string, error) {
	var matched bool
	var matchedObjects []string
	err := c.S3Session.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucketName,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		for _, obj := range p.Contents {
			matched, _ = regexp.MatchString(regex, *obj.Key)
			if matched {
				matchedObjects = append(matchedObjects, *obj.Key)
			}
		}
		return true
	})
	if err != nil {
		klog.Infof("failed to list objects", err)
		return nil, err
	}
	return matchedObjects, err
}

//Func CheckBucketLocationConstraint will verify the existence of the bucket in the particular locationConstraint
func (c *S3Client) CheckBucketLocationConstraint(bucketName string, bucketLocationConstraint string) (bool, error) {

	getParams := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}

	result, err := c.S3Session.GetBucketLocation(getParams)
	if err != nil {
		klog.Infof("Unable to get bucket location %v\n", err)
		return false, err
	}

	if *result.LocationConstraint == bucketLocationConstraint {
		return true, nil
	}
	return false, errors.New("bucket location constraint doesn't match")
}

func (c *S3Client) CheckIfObjectExists(bucketName, objectName string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	_, err := c.S3Session.GetObject(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				klog.Infof("Object %s not found in %s bucket", objectName, bucketName)
				return false, nil
			}
		}
		return false, fmt.Errorf("unknown error occurred, %v", err)
	}
	return true, nil
}

//To create a new bucket in the provided instance
func (c *S3Client) CreateBucket(bucketName string) error {
	_, err := c.S3Session.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName), // New Bucket Name
	})
	if err != nil {
		klog.Errorf("Unable to create bucket %q, %v", bucketName, err)
		return err
	}
	// Wait until bucket is created before finishing
	klog.Infof("Waiting for bucket %q to be created...\n", bucketName)

	err = c.S3Session.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

//To copy the object from src bucket to target bucket
func (c *S3Client) CopyObjectToBucket(srcBucketName string, destBucketName string, objectName string) error {
	copyParams := s3.CopyObjectInput{
		Bucket:     aws.String(destBucketName),
		CopySource: aws.String(srcBucketName + "/" + objectName),
		Key:        aws.String(objectName),
	}
	_, err := c.S3Session.CopyObject(&copyParams)
	if err != nil {
		klog.Errorf("Unable to copy object %s from bucket %s, to bucket %s Error: %v", objectName, srcBucketName, destBucketName, err)
		return err
	}

	klog.Infof("Copy successful for object: %s from bucket: %s to bucket: %s", objectName, srcBucketName, destBucketName)
	return err
}

//To upload a object to S3 bucket
func (c *S3Client) UploadObject(fileName, objectName, bucketName string) error {
	klog.Infof("uploading the file %s\n", fileName)
	//Read the content of the file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("err opening file %s: %s", fileName, err)
	}
	defer file.Close()

	// Create an uploader with S3 client
	uploader := s3manager.NewUploaderWithClient(c.S3Session, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024
	})

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
		Body:   file,
	}

	// Perform an upload.
	startTime := time.Now()
	result, err := uploader.Upload(upParams)
	if err != nil {
		return err
	}
	klog.Infof("Upload completed successfully in %f seconds to location %s\n", time.Since(startTime).Seconds(), result.Location)
	return err
}
