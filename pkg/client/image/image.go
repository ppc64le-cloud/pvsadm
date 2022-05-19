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

package image

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/ppc64le-cloud/pvsadm/pkg"
)

type Client struct {
	session    *ibmpisession.IBMPISession
	client     *instance.IBMPIImageClient
	jobclient  *instance.IBMPIJobClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	c := &Client{
		session:    sess,
		instanceID: powerinstanceid,
	}
	c.client = instance.NewIBMPIImageClient(context.Background(), sess, powerinstanceid)
	c.jobclient = instance.NewIBMPIJobClient(context.Background(), sess, powerinstanceid)
	return c
}

func (c *Client) Get(id string) (*models.Image, error) {
	return c.client.Get(id)
}

func (c *Client) GetAll() (*models.Images, error) {
	return c.client.GetAll()
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id)
}

func (c *Client) CreateCosImage(body models.CreateCosImageImportJob) (*models.JobReference, error) {
	return c.client.CreateCosImage(&body)
}

//func ImportImage imports image from S3 Instance
func (c *Client) ImportImage(imageName, s3Filename, region, accessKey, secretKey, bucketName, storageType, bucketAccess string) (*models.JobReference, error) {

	var body = models.CreateCosImageImportJob{
		ImageName:     &imageName,
		ImageFilename: &s3Filename,
		Region:        &region,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
		BucketName:    &bucketName,
		StorageType:   storageType,
		BucketAccess:  &bucketAccess,
	}

	jobRef, err := c.CreateCosImage(body)
	if err != nil {
		return nil, err
	}

	return jobRef, nil
}

func (c *Client) GetAllPurgeable(before, since time.Duration, expr string) ([]*models.ImageReference, error) {
	images, err := c.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of instances: %v", err)
	}

	var candidates []*models.ImageReference
	for _, image := range images.Images {
		if expr != "" {
			if r, _ := regexp.Compile(expr); !r.MatchString(*image.Name) {
				continue
			}
		}
		if !pkg.IsPurgeable(time.Time(*image.CreationDate), before, since) {
			continue
		}
		candidates = append(candidates, image)
	}
	return candidates, nil
}

func (c *Client) GetImageByName(imageName string) (*models.ImageReference, error) {
	images, err := c.GetAll()
	if err != nil {
		return nil, err
	}
	for _, img := range images.Images {
		if *img.Name == imageName {
			return img, nil
		}
	}
	return nil, nil
}
