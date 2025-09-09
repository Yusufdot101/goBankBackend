package validator

import (
	"regexp"
	"testing"
)

func TestAddError(t *testing.T) {
	tests := []struct {
		name           string
		setupValidator func(*Validator)
		key            string
		message        string
		finalMessage   string
		length         int
	}{
		{
			name:           "new key and value",
			setupValidator: func(v *Validator) { v.Errors = map[string]string{} }, // empty the prev
			key:            "some key",
			message:        "some message",
			finalMessage:   "some message",
			length:         1,
		},
		{
			name:           "duplicate",
			setupValidator: func(v *Validator) {}, // we don't clean out the prev key
			key:            "some key",
			message:        "some other message",
			finalMessage:   "some message",
			length:         1,
		},
	}

	v := New()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupValidator(v)
			v.AddError(tc.key, tc.message)
			if len(v.Errors) != tc.length {
				t.Fatalf("expected length %d, got %d", tc.length, len(v.Errors))
			}
			if _, exists := v.Errors[tc.key]; !exists {
				t.Fatal("key not found")
			}
			if v.Errors[tc.key] != tc.finalMessage {
				t.Fatalf("expected final message %s, got %s", tc.message, v.Errors[tc.key])
			}
		})
	}
}

func TestCheckAddError(t *testing.T) {
	tests := []struct {
		name           string
		setupValidator func(*Validator)
		key            string
		message        string
		condition      bool
		length         int
	}{
		{
			name:           "valid condition",
			setupValidator: func(v *Validator) { v.Errors = map[string]string{} },
			key:            "result",
			message:        "1 + 1 = 1",
			condition:      1+1 == 2,
			length:         0,
		},
		{
			name:           "invalid condition",
			setupValidator: func(v *Validator) { v.Errors = map[string]string{} },
			key:            "result",
			message:        "1 != 2",
			condition:      1 == 2,
			length:         1,
		},
	}

	v := New()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupValidator(v)
			v.CheckAddError(tc.condition, tc.key, tc.message)
			if len(v.Errors) != tc.length {
				t.Fatalf("expected length %d, got %d", tc.length, len(v.Errors))
			}

			// if its empty, no need to check for the key's existance
			if tc.length == 0 {
				return
			}

			if _, exists := v.Errors[tc.key]; !exists {
				t.Fatal("key not found")
			}
		})
	}
}

func TestValueInList(t *testing.T) {
	tests := []struct {
		name        string
		list        []string
		value       string
		valueInList bool
	}{
		{
			name:        "value in the list",
			list:        []string{"a", "b", "c"},
			value:       "a",
			valueInList: true,
		},
		{
			name:        "value not in the list",
			list:        []string{"a", "b", "c"},
			value:       "d",
			valueInList: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valueFound := ValueInList(tc.value, tc.list...)
			if valueFound != tc.valueInList {
				t.Fatalf("expected found=%v, got found=%v", tc.valueInList, valueFound)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name         string
		regexPattern *regexp.Regexp
		value        string
		matches      bool
	}{
		{
			name:         "pattern matches",
			regexPattern: regexp.MustCompile("^[a-z]{5,}$"), // 5 or more lower case letters
			value:        "hello",
			matches:      true,
		},
		{
			name:         "pattern doesn't match",
			regexPattern: regexp.MustCompile("^[a-z]?$"), // zero or one lower case letter
			value:        "some text",
			matches:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := Matches(tc.value, tc.regexPattern)
			if matches != tc.matches {
				t.Fatalf("expected matches=%v, got matches=%v", tc.matches, matches)
			}
		})
	}
}
