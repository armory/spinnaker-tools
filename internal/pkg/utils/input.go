package utils

import (
	"github.com/manifoldco/promptui"
)

func PromptUntilValid(prompt promptui.Prompt, verbose bool) (string, error) {
	for {
		r, err := prompt.Run()
		if err == nil {
			return r, nil
		}
		if err == promptui.ErrInterrupt || err == promptui.ErrAbort || err == promptui.ErrEOF {
			return "", err
		}
	}
}
