package utils

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"k8s.io/klog/v2"
)

func AskYesOrNo(message string) bool {
	validate := func(input string) error {
		if input != "yes" && input != "no" {
			return fmt.Errorf("only yes/no option supported")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    message + "[yes/no]",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		klog.Fatalf("Prompt failed %v\n", err)
	}

	return result == "yes"
}
