package image

import (
	"fmt"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"time"
)

type Client struct {
	client     *instance.IBMPIImageClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	c := &Client{
		instanceID: powerinstanceid,
	}
	c.client = instance.NewIBMPIImageClient(sess, powerinstanceid)
	return c
}

func (c *Client) Get(id string) (*models.Image, error) {
	return c.client.Get(id, c.instanceID)
}

func (c *Client) GetAll() (*models.Images, error) {
	return c.client.GetAll(c.instanceID)
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id, c.instanceID)
}

func (c *Client) GetAllPurgeable(before, since time.Duration) ([]*models.ImageReference, error) {
	images, err := c.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of instances: %v", err)
	}

	var candidates []*models.ImageReference
	for _, image := range images.Images {
		if !pkg.IsPurgeable(time.Time(*image.CreationDate), before, since) {
			continue
		}
		candidates = append(candidates, image)
	}
	return candidates, nil
}
