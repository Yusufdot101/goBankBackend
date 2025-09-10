package permission

import (
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func TestIncludes(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		permissions []Permission
		wantOutput  bool
	}{
		{
			name:        "valid",
			code:        "code1",
			permissions: []Permission{"code1", "code2"},
			wantOutput:  true,
		},
		{
			name:        "not includes",
			code:        "code1",
			permissions: []Permission{},
			wantOutput:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			includes := Includes(tc.permissions, tc.code)
			if includes != tc.wantOutput {
				t.Fatalf("expected output=%v, got output=%v", tc.wantOutput, includes)
			}
		})
	}
}

func TestValidateCode(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantValid      bool
		expectedErrMsg map[string]string
	}{
		{
			name:      "valid",
			code:      "ADMIN",
			wantValid: true,
		},
		{
			name:      "unsafe code",
			code:      "super-admin",
			wantValid: false,
			expectedErrMsg: map[string]string{
				"code": "invalid",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := validator.New()
			ValidateCode(v, tc.code)
			if v.IsValid() != tc.wantValid {
				t.Fatalf("expected valid=%v, got valid=%v", tc.wantValid, v.IsValid())
			} else if v.IsValid() {
				return
			}

			for key, val := range tc.expectedErrMsg {
				if v.Errors[key] != val {
					t.Errorf(
						"expected message=%s for key=%v, got message=%s", val, key, v.Errors[key],
					)
				}
			}
		})
	}
}
