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

package validate

import (
	"fmt"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"k8s.io/klog/v2"
)

var rules []Rule

type Rule interface {
	Verify() error
	Hint() string
	String() string
}

func AddRule(r Rule) {
	rules = append(rules, r)
}

func Validate() error {
	for _, rule := range rules {
		ruleStr := rule.String()
		klog.Infof("Checking: %s", ruleStr)
		if utils.Contains(pkg.ImageCMDOptions.PreflightSkip, ruleStr) {
			klog.Info("SKIPPED!")
			continue
		}
		err := rule.Verify()
		if err != nil {
			return fmt.Errorf("check failed: %v \nHint: %v", err, rule.Hint())
		}
	}
	return nil
}
