package client

import (
	"errors"
	"os"
)

const DefaultEnv = "prod"

var EnvironmentNotFound = errors.New("environment not found")

var Environments = map[string]map[string]string{
	"test": {
		"TPEndpoint": "https://iam.test.cloud.ibm.com",
		"RCEndpoint": "https://resource-controller.test.cloud.ibm.com",
		"PIEndpoint": "power-iaas.test.cloud.ibm.com",
	},
	"prod": {
		"TPEndpoint": "https://iam.cloud.ibm.com",
		"RCEndpoint": "https://resource-controller.cloud.ibm.com",
		"PIEndpoint": "power-iaas.cloud.ibm.com",
	},
}

func ListEnvironments() (keys []string) {
	for k := range Environments {
		keys = append(keys, k)
	}
	return
}

func GetEnvironment(env string) (map[string]string, error) {
	if _, ok := Environments[env]; !ok {
		return nil, EnvironmentNotFound
	}
	return Environments[env], nil
}

func NewPVMClientWithEnv(c *Client, instanceID, instanceName, env string) (*PVMClient, error) {
	e, err := GetEnvironment(env)
	if err != nil {
		return nil, err
	}
	return NewPVMClient(c, instanceID, instanceName, e["PIEndpoint"])
}

func NewClientWithEnv(apikey, env string, debug bool) (*Client, error) {
	e, err := GetEnvironment(env)
	if err != nil {
		return nil, err
	}
	os.Setenv("IBMCLOUD_RESOURCE_CONTROLLER_API_ENDPOINT", e["RCEndpoint"])
	return NewClient(apikey, e["TPEndpoint"], debug)
}
