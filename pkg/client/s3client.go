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
	"io"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
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

// Func NewS3ClientWithKeys accepts apikey, accesskey, secretkey of the bucket and returns the s3 client
// to perform operations like upload, delete, etc.
func NewS3ClientWithKeys(accesskey, secretkey, region string) (s3client *S3Client, err error) {

	s3client = &S3Client{
		SvcEndpoint:  fmt.Sprintf("https://s3.%s.cloud-object-storage.appdomain.cloud", region),
		StorageClass: fmt.Sprintf("%s-standard", region),
	}
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

// NewS3Client accepts apikey, instanceid of the IBM COS instance and return the s3 client
// to perform different s3 operations like upload, delete etc.,
func NewS3Client(c *Client, instanceName, region string) (s3client *S3Client, err error) {
	s3client = &S3Client{}

	listServiceInstanceOptions := &resourcecontrollerv2.ListResourceInstancesOptions{
		Type: ptr.To(serviceInstance),
		Name: ptr.To(instanceName),
	}

	s3Services, _, err := c.ResourceControllerClient.ListResourceInstances(listServiceInstanceOptions)
	if err != nil {
		return s3client, fmt.Errorf("failed to list the resource instances, err: %v", err)
	}
	found := false
	for _, svc := range s3Services.Resources {
		klog.V(3).Infof("Service ID: %s, region_id: %s, Name: %s", *svc.ID, *svc.RegionID, *svc.Name)
		klog.V(3).Infof("crn: %v", *svc.CRN)
		if *svc.Name == instanceName {
			s3client.InstanceID = *svc.GUID
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("instance: %s not found", instanceName)
	}

	if pkg.Options.APIKey == "" {
		s3client.ApiKey = os.Getenv("IBMCLOUD_APIKEY")
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

// CheckBucketExists will verify for the existence of the bucket in the particular account
func (c *S3Client) CheckBucketExists(bucketName string) (bool, error) {
	result, err := c.S3Session.ListBuckets(nil)
	if err != nil {
		klog.Errorf("unable to list buckets, err: %v", err)
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
		klog.Errorf("failed to list objects, err: %v", err)
		return nil, err
	}
	return matchedObjects, nil
}

// Func CheckBucketLocationConstraint will verify the existence of the bucket in the particular locationConstraint
func (c *S3Client) CheckBucketLocationConstraint(bucketName string, bucketLocationConstraint string) (bool, error) {

	getParams := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}

	result, err := c.S3Session.GetBucketLocation(getParams)
	if err != nil {
		klog.Errorf("unable to get bucket location, err: %v", err)
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
				klog.Errorf("object %s not found in %s bucket", objectName, bucketName)
				return false, nil
			}
		}
		return false, fmt.Errorf("unknown error occurred, err: %v", err)
	}
	return true, nil
}

// CreateBucket creates a new bucket in the provided instance
func (c *S3Client) CreateBucket(bucketName string) error {
	_, err := c.S3Session.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName), // New Bucket Name
	})
	if err != nil {
		klog.Errorf("unable to create bucket %q, err: %v", bucketName, err)
		return err
	}
	// Wait until bucket is created before finishing
	klog.Infof("Waiting for bucket %s to be created.", bucketName)

	err = c.S3Session.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// CopyObjectToBucket copies the object from src bucket to target bucket
func (c *S3Client) CopyObjectToBucket(srcBucketName string, destBucketName string, objectName string) error {
	copyParams := s3.CopyObjectInput{
		Bucket:     aws.String(destBucketName),
		CopySource: aws.String(srcBucketName + "/" + objectName),
		Key:        aws.String(objectName),
	}
	if _, err := c.S3Session.CopyObject(&copyParams); err != nil {
		klog.Errorf("unable to copy object %s from bucket %s, to bucket %s, err: %v", objectName, srcBucketName, destBucketName, err)
		return err
	}

	klog.Infof("Copy successful for object: %s from bucket: %s to bucket: %s", objectName, srcBucketName, destBucketName)
	return nil
}

type CustomReader struct {
	fp              *os.File
	size            int64
	read            int64
	mux             sync.Mutex
	progresstracker *ProgressTracker
	signMap         map[int64]struct{}
}

type ProgressTracker struct {
	progress *mpb.Progress
	bar      *mpb.Bar
	isBarSet bool
	counter  *formattedCounter
}

type formattedCounter struct {
	read  *int64
	total int64
}

func (f *formattedCounter) Decor(stat decor.Statistics) (string, int) {
	str := fmt.Sprintf("%s/%s", formatBytes(*f.read), formatBytes(f.total))
	return str, len(str)
}

func (f *formattedCounter) Format(string) (string, int) {
	return "", 0
}

func (f *formattedCounter) Sync() (chan int, bool) {
	return nil, false
}

func (r *CustomReader) Read(p []byte) (int, error) {
	return r.fp.Read(p)
}

func (r *CustomReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.fp.ReadAt(p, off)
	if err != nil {
		if err == io.EOF {
			return n, nil
		}
		return n, err
	}
	r.mux.Lock()
	if _, ok := r.signMap[off]; ok {
		r.read += int64(n)
		r.progresstracker.counter.read = &r.read
	} else {
		r.signMap[off] = struct{}{}
	}
	r.progresstracker.bar.SetCurrent(r.read)
	r.mux.Unlock()
	return n, nil
}

// Format the bytes to a human-readable string
func formatBytes(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}

func (r *CustomReader) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}

// To upload a object to S3 bucket
func (c *S3Client) UploadObject(fileName, objectName, bucketName string) error {
	klog.Infof("Uploading the file %s", fileName)
	// Read the content of the file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("err opening file %s, err: %s", fileName, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %v, err: %v", fileName, err)
	}

	// Create the custom reader for file reading and progress tracking
	reader := &CustomReader{
		fp:      file,
		size:    fileInfo.Size(),
		signMap: map[int64]struct{}{},
	}

	// Initialize progress tracker
	progressTracker := &ProgressTracker{
		progress: mpb.New(),
		counter:  &formattedCounter{read: new(int64), total: reader.size},
	}

	bar := progressTracker.progress.AddBar(reader.size,
		mpb.PrependDecorators(
			decor.Name("Uploading: ", decor.WC{W: 15}),
			progressTracker.counter,
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)
	reader.progresstracker = progressTracker
	reader.progresstracker.bar = bar

	// Create an uploader with S3 client
	uploader := s3manager.NewUploaderWithClient(c.S3Session, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024
	})

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
		Body:   reader,
	}

	// Perform an upload
	startTime := time.Now()
	result, err := uploader.Upload(upParams)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	progressTracker.progress.Wait()

	klog.Infof("Upload completed successfully in %s to location %s", time.Since(startTime).Round(time.Second), result.Location)
	return nil
}
