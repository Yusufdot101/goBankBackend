package permission

import (
	"errors"
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type MockRepo struct {
	AllForUserResult []Permission
	AllForUserErr    error

	GrantErr  error
	RevokeErr error
	DeleteErr error
	InsertErr error
}

func (r *MockRepo) AllForUser(userID int64) ([]Permission, error) {
	if r.AllForUserErr != nil {
		return nil, r.AllForUserErr
	}
	return r.AllForUserResult, nil
}

func (r *MockRepo) Grant(userID int64, code ...string) error {
	return r.GrantErr
}

func (r *MockRepo) Revoke(userID int64, code ...string) error {
	return r.RevokeErr
}

func (r *MockRepo) Delete(code ...string) error {
	return r.DeleteErr
}

func (r *MockRepo) Insert(code Permission) error {
	return r.InsertErr
}

type MockUserService struct {
	GetUserResult *user.User
	GetUserErr    error
}

func (us *MockUserService) GetUser(userID int64) (*user.User, error) {
	if us.GetUserErr != nil {
		return nil, us.GetUserErr
	}
	return us.GetUserResult, nil
}

func TestUserHas(t *testing.T) {
	mockPermissions := []Permission{"ADMIN", "SUPERUSER"}
	tests := []struct {
		name        string
		setupRepo   func(*MockRepo)
		code        string
		wantOutput  bool
		expectedErr error
	}{
		{
			name: "valid",
			setupRepo: func(r *MockRepo) {
				r.AllForUserResult = mockPermissions
			},
			code:       "ADMIN",
			wantOutput: true,
		},
		{
			name: "user doesn't have permission",
			setupRepo: func(r *MockRepo) {
				r.AllForUserResult = mockPermissions
			},
			code:       "APPROVE_LOANS",
			wantOutput: false,
		},
		{
			name: "AllForUser failure",
			setupRepo: func(r *MockRepo) {
				r.AllForUserErr = user.ErrNoRecord
			},
			code:        "DELETE_LOANS",
			wantOutput:  false,
			expectedErr: user.ErrNoRecord,
		},
		{
			name:        "unsafe code",
			setupRepo:   func(r *MockRepo) {},
			code:        "code3",
			wantOutput:  false,
			expectedErr: validator.ErrFailedValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tc.setupRepo(repo)

			svc := Service{
				Repo: repo,
			}

			v := validator.New()
			has, gotErr := svc.UserHas(v, &user.User{}, tc.code)
			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}

			if has != tc.wantOutput {
				t.Errorf("expected has=%v, got has=%v", tc.wantOutput, has)
			}
		})
	}
}

func TestGrantUser(t *testing.T) {
	mockUser := &user.User{
		ID:    1,
		Name:  "yusuf",
		Email: "ym@gmail.com",
	}
	tests := []struct {
		name             string
		setupRepo        func(*MockRepo)
		setupUserService func(*MockUserService)
		code             string
		expectedErr      error
	}{
		{
			name:      "valid",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			code: "ADMIN",
		},
		{
			name:             "unsafe code",
			setupRepo:        func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {},
			code:             "random",
			expectedErr:      validator.ErrFailedValidation,
		},
		{
			name:      "GetUser failure",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserErr = errors.New("db GetUser error")
			},
			code:        "ADMIN",
			expectedErr: errors.New("db GetUser error"),
		},
		{
			name: "Grant failure",
			setupRepo: func(r *MockRepo) {
				r.GrantErr = errors.New("db Grant error")
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			code:        "ADMIN",
			expectedErr: errors.New("db Grant error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			userSvc := &MockUserService{}
			tc.setupRepo(repo)
			tc.setupUserService(userSvc)

			svc := Service{
				Repo:        repo,
				UserService: userSvc,
			}

			v := validator.New()
			gotErr := svc.GrantUser(v, mockUser.ID, tc.code)
			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}
		})
	}
}

func TestRevokeFromUser(t *testing.T) {
	mockUser := &user.User{
		ID:    1,
		Name:  "yusuf",
		Email: "ym@gmail.com",
	}
	tests := []struct {
		name             string
		setupRepo        func(*MockRepo)
		setupUserService func(*MockUserService)
		code             string
		expectedErr      error
	}{
		{
			name:      "valid",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			code: "ADMIN",
		},
		{
			name:             "unsafe code",
			setupRepo:        func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {},
			code:             "random",
			expectedErr:      validator.ErrFailedValidation,
		},
		{
			name:      "GetUser failure",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserErr = errors.New("db GetUser error")
			},
			code:        "ADMIN",
			expectedErr: errors.New("db GetUser error"),
		},
		{
			name: "Revoke failure",
			setupRepo: func(r *MockRepo) {
				r.RevokeErr = errors.New("db Revoke error")
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			code:        "ADMIN",
			expectedErr: errors.New("db Revoke error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			userSvc := &MockUserService{}
			tc.setupRepo(repo)
			tc.setupUserService(userSvc)

			svc := Service{
				Repo:        repo,
				UserService: userSvc,
			}

			v := validator.New()
			gotErr := svc.RevokeFromUser(v, mockUser.ID, tc.code)
			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}
		})
	}
}

func TestAddNewPermission(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*MockRepo)
		code        string
		expectedErr error
	}{
		{
			name:      "valid",
			setupRepo: func(mr *MockRepo) {},
			code:      "ADMIN",
		},
		{
			name:        "unsafe code",
			setupRepo:   func(mr *MockRepo) {},
			code:        "random",
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name: "Insert failure",
			setupRepo: func(r *MockRepo) {
				r.InsertErr = errors.New("db Insert error")
			},
			code:        "ADMIN",
			expectedErr: errors.New("db Insert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tc.setupRepo(repo)

			svc := Service{
				Repo: repo,
			}

			v := validator.New()
			gotErr := svc.AddNewPermission(v, tc.code)
			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}
		})
	}
}
