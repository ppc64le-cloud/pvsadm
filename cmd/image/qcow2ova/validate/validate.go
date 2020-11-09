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
		klog.Infof("Checking: %s\n", ruleStr)
		if utils.Contains(pkg.ImageCMDOptions.PreflightSkip, ruleStr) {
			klog.Infof("SKIPPED!")
			continue
		}
		err := rule.Verify()
		if err != nil {
			return fmt.Errorf("check failed: %v \nHint: %v", err, rule.Hint())
		}
	}
	return nil
}
