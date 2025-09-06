package permission

import (
	"slices"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Permission string

func Includes(permissions []Permission, code ...string) bool {
	for _, value := range code {
		if slices.Contains(permissions, Permission(value)) {
			return true
		}
	}

	return false
}

func ValidateCode(v *validator.Validator, code string) {
	v.CheckAddError(code != "", "code", "must be provided")
	safePermissions := []string{
		"APPROVE_LOANS",
		"DELETE_LOANS",
		"ADMIN",
		"SUPERUSER",
	}
	v.CheckAddError(validator.ValueInList(code, safePermissions...), "code", "invalid")
}
