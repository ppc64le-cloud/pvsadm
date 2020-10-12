package instance

import (
	"fmt"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"regexp"
	"time"
)

type Client struct {
	client     *instance.IBMPIInstanceClient
	instanceID string
}

func NewClient(sess *ibmpisession.IBMPISession, powerinstanceid string) *Client {
	c := &Client{
		instanceID: powerinstanceid,
	}
	c.client = instance.NewIBMPIInstanceClient(sess, powerinstanceid)
	return c
}

func (c *Client) Get(id string) (*models.PVMInstance, error) {
	return c.client.Get(id, c.instanceID, pkg.TIMEOUT)
}

func (c *Client) GetAll() (*models.PVMInstances, error) {
	return c.client.GetAll(c.instanceID, pkg.TIMEOUT)
}

func (c *Client) Delete(id string) error {
	return c.client.Delete(id, c.instanceID, pkg.TIMEOUT)
}

func (c *Client) GetAllPurgeable(before, since time.Duration, expr string) ([]*models.PVMInstanceReference, error) {
	instances, err := c.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of instances: %v", err)
	}

	var candidates []*models.PVMInstanceReference
	for _, ins := range instances.PvmInstances {
		if expr != "" {
			if r, _ := regexp.Compile(expr); !r.MatchString(*ins.ServerName) {
				continue
			}
		}
		if !pkg.IsPurgeable(time.Time(ins.CreationDate), before, since) {
			continue
		}
		candidates = append(candidates, ins)
	}
	return candidates, nil
}
