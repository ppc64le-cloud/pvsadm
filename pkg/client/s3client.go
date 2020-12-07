package client

import (
	"fmt"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev2/controllerv2"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/IBM/ibm-cos-sdk-go/service/s3/s3manager"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"time"
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

//Func NewS3Client accepts apikey, instanceid of the IBM COS instance and return the s3 client
//to perform different s3 operations like upload, delete etc.,
func NewS3Client(c *Client, instanceName, region string) (s3client *S3Client, err error) {
	s3client = &S3Client{}
	var instanceID string
	svcs, err := c.ResourceClient.ListInstances(controllerv2.ServiceInstanceQuery{
		Type: "service_instance",
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

func (c *S3Client) CheckIfObjectExists(bucketName, objectName string) bool {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	_, err := c.S3Session.GetObject(input)
	if err != nil {
		return false
	}
	return true
}

//To create a new bucket in the provided instance
func (c *S3Client) CreateBucket(bucketName string) error {
	_, err := c.S3Session.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName), // New Bucket Name
	})
	if err != nil {
		return fmt.Errorf("Unable to create bucket %q, %v", bucketName, err)
	}
	// Wait until bucket is created before finishing
	klog.Infof("Waiting for bucket %q to be created...\n", bucketName)

	err = c.S3Session.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

//To upload a object to S3 bucket
func (c *S3Client) UploadObject(object, bucketName string) error {
	klog.Infof("uploading the object %s\n", object)
	//Read the content of the object
	file, err := os.Open(object)
	if err != nil {
		return fmt.Errorf("err opening file %s: %s", object, err)
	}
	defer file.Close()

	// Create an uploader with S3 client
	uploader := s3manager.NewUploaderWithClient(c.S3Session, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024
	})

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Base(object)),
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
