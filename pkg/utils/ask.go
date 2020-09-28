package utils

import (
	"github.com/manifoldco/promptui"
	"k8s.io/klog/v2"
)

func AskYesOrNo(message string) bool {
	prompt := promptui.Select{
		Label: message,
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		klog.Fatalf("Prompt failed %v", err)
	}

	return result == "Yes"
}
