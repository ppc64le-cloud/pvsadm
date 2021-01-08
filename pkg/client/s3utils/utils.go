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

package s3utils

import (
	"bufio"
	"fmt"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev2/controllerv2"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"k8s.io/klog/v2"
	"os"
	"strconv"
	"strings"
)

//Func GetInstances, list all available instances of particular servicetype
func GetInstances(c *client.Client, serviceType string) (map[string]string, error) {
	svcs, err := c.ResourceClient.ListInstances(controllerv2.ServiceInstanceQuery{
		Type: "service_instance",
	})

	if err != nil {
		return nil, err
	}

	instances := make(map[string]string)

	for _, svc := range svcs {
		if svc.Crn.ServiceName == serviceType {
			instances[svc.Name] = svc.Guid
		}
	}
	return instances, nil
}

//Func AskYesOrNo, will read the user's input and returns bool
func AskYesOrNo(promptMsg string, tries int) bool {
	r := bufio.NewReader(os.Stdin)
	for ; tries > 0; tries-- {
		fmt.Printf("%s [y/n]: ", promptMsg)

		res, err := r.ReadString('\n')
		if err != nil {
			klog.Fatal(err)
		}
		response := strings.ToLower(strings.TrimSpace(res))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		} else {
			fmt.Printf("Enter a valid input:[yes/no]\n")
			continue
		}
	}

	return false
}

//Func SelectCosInstance, reads the selected instance from the user
func SelectCosInstance(availableCosInstance, tries int) int {
	reader := bufio.NewReader(os.Stdin)
	for ; tries > 0; tries-- {
		fmt.Printf("Enter a number:")
		inputstr, err := reader.ReadString('\n')
		if err != nil {
			klog.Fatal(err)
		}
		input, err := strconv.ParseInt(strings.TrimSpace(inputstr), 10, 64)
		if err != nil {
			klog.Fatal(err)
		}

		if int(input) >= availableCosInstance {
			fmt.Printf("Please select a valid COS Instance\n")
			continue
		}
		return int(input)
	}
	return -1
}

//Func ReadInstanceNameFromUser, Read the Instance name from user to create a new instance
func ReadInstanceNameFromUser() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Provide Name of the cos-instance:")
	inputstr, err := reader.ReadString('\n')
	if err != nil {
		klog.Fatal(err)
	}
	return strings.TrimSpace(inputstr)
}
