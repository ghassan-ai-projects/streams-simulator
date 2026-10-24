package model

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

const (
	MaxPlanNameChars    = 80
	MaxPlanPurposeChars = 500
)

func CheckPlanText(name, purpose string) error {
	if err := checkPlainText("name", name, MaxPlanNameChars); err != nil {
		return err
	}
	return checkPlainText("purpose", purpose, MaxPlanPurposeChars)
}

func checkPlainText(field, text string, limit int) error {
	if !utf8.ValidString(text) {
		return fmt.Errorf("%s must be valid UTF-8 text", field)
	}
	if utf8.RuneCountInString(text) > limit {
		return fmt.Errorf("%s must be at most %d characters", field, limit)
	}
	for _, r := range text {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s must be plain text without control characters", field)
		}
	}
	return nil
}
