package permission

import (
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Repo interface {
	AllForUser(userID int64) ([]Permission, error)
	Grant(userID int64, code ...string) error
	Revoke(userID int64, code ...string) error
	Delete(code ...string) error
	Insert(code Permission) error
}

type UserService interface {
	GetUser(userID int64) (*user.User, error)
}

type Service struct {
	Repo        Repo
	UserService UserService
}

func (s *Service) UserHas(v *validator.Validator, u *user.User, code string) (bool, error) {
	if ValidateCode(v, code); !v.IsValid() {
		return false, validator.ErrFailedValidation
	}

	userPermissions, err := s.Repo.AllForUser(u.ID)
	if err != nil {
		return false, err
	}

	return Includes(userPermissions, code), nil
}

func (s *Service) UserAllPermissions(userID int64) ([]Permission, error) {
	allPermissions, err := s.Repo.AllForUser(userID)
	if err != nil {
		return nil, err
	}

	return allPermissions, nil
}

func (s *Service) GrantUser(v *validator.Validator, userID int64, code string) error {
	if ValidateCode(v, code); !v.IsValid() {
		return validator.ErrFailedValidation
	}

	// verify the user exists
	u, err := s.UserService.GetUser(userID)
	if err != nil {
		return err
	}

	return s.Repo.Grant(u.ID, code)
}

func (s *Service) RevokeFromUser(v *validator.Validator, userID int64, code string) error {
	if ValidateCode(v, code); !v.IsValid() {
		return validator.ErrFailedValidation
	}
	// verify the user exists
	u, err := s.UserService.GetUser(userID)
	if err != nil {
		return err
	}

	return s.Repo.Revoke(u.ID, code)
}

func (s *Service) DeletePermission(code string) error {
	return s.Repo.Delete(code)
}

func (s *Service) AddNewPermission(v *validator.Validator, code string) error {
	if ValidateCode(v, code); !v.IsValid() {
		return validator.ErrFailedValidation
	}

	err := s.Repo.Insert(Permission(code))
	if err != nil {
		return err
	}

	return nil
}
