// Copyright 2022 IBM Corp
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

package utils

import (
	"errors"

	"k8s.io/klog/v2"

	"github.com/charmbracelet/huh"
	access "github.com/charmbracelet/huh/accessibility"
)

func SelectItem(msg string, instances []string) (string, error) {
	var choice string
	err :=
		huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().Title(msg).
					Options(huh.NewOptions(instances...)...).Value(&choice))).Run()

	if err != nil {
		klog.Errorf("couldn't process the inputs: %v", err)
		return "", err
	}
	return choice, nil
}

func AskConfirmation(message string) bool {
	var confirm bool
	err :=
		huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(message).
					Affirmative("Yes").
					Negative("No").
					Value(&confirm))).Run()
	if err != nil {
		klog.Fatalf("couldn't process inputs: %v", err)

	}
	return confirm
}

func ReadUserInput(message string) string {
	validateInput := func(data string) error {
		if data == "" {
			return errors.New("input cannot be empty, please retry")
		}
		return nil
	}

	return access.PromptString(message, validateInput)
}
