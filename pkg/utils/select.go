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

import "github.com/AlecAivazis/survey/v2"

func SelectItem(msg string, instances []string) string {
	instance := ""
	prompt := &survey.Select{
		Message: msg,
		Options: instances,
	}
	survey.AskOne(prompt, &instance)
	return instance
}

func AskConfirmation(message string) bool {
	result := false
	prompt := &survey.Confirm{
		Message: message,
	}
	survey.AskOne(prompt, &result)
	return result
}

func ReadUserInput(message string) string {
	name := ""
	prompt := &survey.Input{
		Message: message,
	}
	survey.AskOne(prompt, &name, survey.WithValidator(survey.Required))
	return name
}

func MultiSelect(msg string, input []string) []string {
	selected := []string{}
	prompt := &survey.MultiSelect{
		Message: msg,
		Options: input,
	}

	survey.AskOne(prompt, &selected)
	return selected
}
