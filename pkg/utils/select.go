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
