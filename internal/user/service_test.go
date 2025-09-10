package user

import (
	"errors"
	"testing"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

// ---MOCKS---

type MockRepo struct {
	InsertErr error

	GetForTokenResult *User
	GetForTokenErr    error

	UpdateTxResult *User
	UpdateTxErr    error
}

func (r *MockRepo) Insert(user *User) error {
	return r.InsertErr
}

func (r *MockRepo) Get(userID int64) (*User, error) {
	return nil, nil
}

func (r *MockRepo) GetByEmail(email string) (*User, error) {
	return nil, nil
}

func (r *MockRepo) GetForToken(tokenPlaintext, scope string) (*User, error) {
	return r.GetForTokenResult, r.GetForTokenErr
}

func (r *MockRepo) UpdateTx(
	userID int64, name, email string, passwordHash []byte, balance float64, activate bool,
) (*User, error) {
	return r.UpdateTxResult, r.UpdateTxErr
}

// ---Mock TokenService---
type MockTokenService struct {
	NewResult *token.Token
	NewErr    error

	DeleteAllErr error
}

func (ts *MockTokenService) New(
	userID int64, timeToLive time.Duration, scope string,
) (*token.Token, error) {
	if ts.NewErr != nil {
		return nil, ts.NewErr
	}

	return ts.NewResult, nil
}

func (ts *MockTokenService) DeleteAllForUser(userID int64, scope string) error {
	return ts.DeleteAllErr
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func(*MockRepo)
		setupTokenSvc func(*MockTokenService)
		input         struct{ Name, Email, Password string }
		expectedErr   error
	}{
		{
			name:      "valid input",
			setupRepo: func(m *MockRepo) {},
			setupTokenSvc: func(ts *MockTokenService) {
				ts.NewResult = &token.Token{Plaintext: "mock-token"}
			},
			input: struct {
				Name     string
				Email    string
				Password string
			}{"yusuf", "a@b.com", "12345678"},
			expectedErr: nil,
		},
		{
			name:          "missing name",
			setupRepo:     func(m *MockRepo) {},
			setupTokenSvc: func(ts *MockTokenService) {},
			input: struct {
				Name     string
				Email    string
				Password string
			}{"", "a@b.com", "12345678"},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name: "duplicate email",
			setupRepo: func(m *MockRepo) {
				m.InsertErr = ErrDuplicateEmail
			},
			setupTokenSvc: func(ts *MockTokenService) {},
			input: struct {
				Name     string
				Email    string
				Password string
			}{"yusuf", "a@b.com", "12345678"},
			expectedErr: ErrDuplicateEmail,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tokenSvc := &MockTokenService{}
			tc.setupRepo(repo)
			tc.setupTokenSvc(tokenSvc)
			svc := &Service{
				Repo:         repo,
				TokenService: tokenSvc,
			}
			v := validator.New()

			user, tkn, err := svc.Register(v, tc.input.Name, tc.input.Email, tc.input.Password)
			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected err %v, got %v", tc.expectedErr, err)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error %s", err)
			}

			if user == nil || tkn == nil {
				t.Fatal("expected user and token to be returned")
			}
		})
	}
}

func TestActivate(t *testing.T) {
	mockUser := &User{
		ID: 1, Name: "yoyo", Email: "a@b.com", Activated: false,
	}

	tests := []struct {
		name          string
		setupRepo     func(*MockRepo)
		setupTokenSvc func(*MockTokenService)
		wantActived   bool
		expectedErr   error
	}{
		{
			name: "valid token",
			setupRepo: func(r *MockRepo) {
				r.GetForTokenResult = mockUser
				r.UpdateTxResult = &User{Activated: true}
			},
			setupTokenSvc: func(ts *MockTokenService) {},
			wantActived:   true,
			expectedErr:   nil,
		},
		{
			name: "invalid token",
			setupRepo: func(r *MockRepo) {
				r.GetForTokenErr = ErrNoRecord
			},
			setupTokenSvc: func(ts *MockTokenService) {},
			wantActived:   false,
			expectedErr:   token.ErrInvaildToken,
		},
		{
			name: "valid token, update user failure",
			setupRepo: func(r *MockRepo) {
				r.GetForTokenResult = &User{Activated: true}
				r.UpdateTxErr = errors.New("db fail")
			},
			setupTokenSvc: func(ts *MockTokenService) {},
			wantActived:   false,
			expectedErr:   errors.New("db fail"),
		},
		{
			name: "valid token, delete token failure",
			setupRepo: func(r *MockRepo) {
				r.GetForTokenResult = mockUser
				r.UpdateTxResult = &User{Activated: true}
			},
			setupTokenSvc: func(ts *MockTokenService) {
				ts.DeleteAllErr = errors.New("db fail")
			},
			wantActived: true,
			expectedErr: errors.New("db fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tokenSvc := &MockTokenService{}
			tc.setupRepo(repo)
			tc.setupTokenSvc(tokenSvc)

			svc := &Service{
				Repo:         repo,
				TokenService: tokenSvc,
			}

			gotUser, gotErr := svc.Activate("token")

			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error: %v", gotErr)
			}

			if gotUser == nil {
				t.Fatal("expected user got nill")
			}
			if gotUser.Activated != tc.wantActived {
				t.Fatalf("expected activate=%v got=%v", tc.wantActived, gotUser.Activated)
			}
		})
	}
}

func TestTransferMoney(t *testing.T) {
	fromUser := &User{
		ID: 1, Name: "yusuf", Email: "a@b.com", AccountBalance: 100,
	}
	toUser := &User{
		ID: 2, Name: "mohamed", Email: "b@a.com", AccountBalance: 50,
	}

	tests := []struct {
		name        string
		amount      float64
		setupRepo   func(*MockRepo)
		expectedErr error
		finalFrom   float64
		finalTo     float64
	}{
		{
			name:   "valid input",
			amount: 10,
			setupRepo: func(r *MockRepo) {
				// after deduct
				r.UpdateTxResult = &User{ID: 1, AccountBalance: 90}
			},
			finalFrom:   90,
			finalTo:     60,
			expectedErr: nil,
		},
		{
			name:   "update failure",
			amount: 10,
			setupRepo: func(r *MockRepo) {
				r.UpdateTxErr = errors.New("db error")
			},
			finalFrom:   100,
			finalTo:     50,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tc.setupRepo(repo)
			svc := Service{
				Repo: repo,
			}

			gotUser, gotErr := svc.TransferMoney(fromUser, toUser, tc.amount)
			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got error %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}

			if gotUser.AccountBalance != tc.finalFrom {
				t.Fatalf(
					"expected balances from=%v, to=%v; got from=%v, to=%v", tc.finalFrom,
					tc.finalTo, fromUser.AccountBalance, toUser.AccountBalance,
				)
			}
		})
	}
}
