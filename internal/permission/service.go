package permission

import (
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Service struct {
	Repo *Repository
}

func (s *Service) UserHas(u *user.User, code string) (bool, error) {
	userPermissions, err := s.Repo.AllForUser(u.ID)
	if err != nil {
		return false, err
	}

	return Includes(userPermissions, code), nil
}

func (s *Service) GrantUser(v *validator.Validator, userID int64, code string) error {
	if ValidateCode(v, code); !v.IsValid() {
		return nil
	}

	userService := user.Service{
		Repo: &user.Repository{DB: s.Repo.DB},
	}
	// verify the user exists
	u, err := userService.Repo.Get(userID)
	if err != nil {
		return err
	}

	return s.Repo.Grant(u.ID, code)
}

func (s *Service) RevokeFromUser(userID int64, code string) error {
	userService := user.Service{
		Repo: &user.Repository{DB: s.Repo.DB},
	}
	// verify the user exists
	u, err := userService.Repo.Get(userID)
	if err != nil {
		return err
	}

	return s.Repo.Revoke(u.ID, code)
}

func (s *Service) Delete(code string) error {
	return s.Repo.Delete(code)
}

func (s *Service) AddNewPermission(v *validator.Validator, code string) error {
	if v.CheckAddError(code != "", "code", "must be given"); !v.IsValid() {
		return nil
	}

	err := s.Repo.Insert(Permission(code))
	if err != nil {
		return err
	}

	return nil
}
