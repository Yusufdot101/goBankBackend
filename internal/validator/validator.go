package validator

import (
	"errors"
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile(
		"^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$",
	)
	ErrFailedValidation = errors.New("failed validation")
)

// Validator will hold the validatior errors
type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0
}

// AddError adds the error message at the key index if the index doesn't exist on the Errors map
func (v *Validator) AddError(key, message string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = message
	}
}

// CheckAddError adds the error message at the key index if the index doesn't exist on the Errors
// map and the condition fails
func (v *Validator) CheckAddError(condition bool, key, message string) {
	if !condition {
		v.AddError(key, message)
	}
}

// ValueInList checks if string value is in list
func ValueInList(value string, list ...string) bool {
	return slices.Contains(list, value)
}

// Matches checks if value string matches the regular expresstion
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
