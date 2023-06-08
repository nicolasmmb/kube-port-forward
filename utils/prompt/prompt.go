package prompt

import (
	"github.com/manifoldco/promptui"
)

func Select(label string, choices []string) (int, string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: choices,
		Size:  10,
	}

	index, result, err := prompt.Run()

	return index, result, err
}

func Confirm(label string, confirm bool, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: confirm,
		Default:   defaultValue,
	}

	result, err := prompt.Run()

	return result, err
}
