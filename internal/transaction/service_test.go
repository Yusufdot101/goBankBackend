package transaction

import (
	"errors"
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type MockRepo struct {
	InsertErr error
}

func (r *MockRepo) Insert(transaction *Transaction) error {
	return r.InsertErr
}

type MockUserService struct {
	GetUserResult *user.User
	GetUserErr    error

	UpdateUserResult *user.User
	UpdateUserErr    error
}

func (us *MockUserService) GetUser(userID int64) (*user.User, error) {
	if us.GetUserErr != nil {
		return nil, us.GetUserErr
	}
	return us.GetUserResult, nil
}

func (us *MockUserService) UpdateUser(
	userID int64, userName, userEmail string, userPasswordHash []byte,
	userAccountBalance float64, userActivated bool,
) (*user.User, error) {
	if us.UpdateUserErr != nil {
		return nil, us.UpdateUserErr
	}
	return us.UpdateUserResult, nil
}

func TestDeposit(t *testing.T) {
	mockUser := &user.User{
		ID:             1,
		Name:           "yusuf",
		Email:          "ym@gmail.com",
		AccountBalance: 0,
	}
	tests := []struct {
		name             string
		setupRepo        func(*MockRepo)
		setupUserService func(*MockUserService)
		input            struct {
			v           *validator.Validator
			userID      int64
			amount      float64
			performedBy string
		}
		expectedErr error
	}{
		{
			name:      "valid",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
		},
		{
			name:      "amount = 0",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 0, performedBy: "yusuf"},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "amount < 0",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: -100, performedBy: "yusuf"},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "GetUser failure",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserErr = user.ErrNoRecord
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
			expectedErr: user.ErrNoRecord,
		},
		{
			name: "Insert error",
			setupRepo: func(r *MockRepo) {
				r.InsertErr = errors.New("db Insert error")
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
			expectedErr: errors.New("db Insert error"),
		},
		{
			name:      "UpdateUser failure",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
				us.UpdateUserErr = errors.New("db UpdateUser error")
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
			expectedErr: errors.New("db UpdateUser error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			userService := &MockUserService{}
			tc.setupRepo(repo)
			tc.setupUserService(userService)

			svc := Service{
				Repo:        repo,
				UserService: userService,
			}

			transaction, gotErr := svc.Deposit(
				tc.input.v, tc.input.userID, tc.input.amount, tc.input.performedBy,
			)

			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}

			depositAction := "DEPOSIT"
			if transaction.Action != depositAction {
				t.Errorf(
					"expected transaction action=%s, got action=%s", depositAction,
					transaction.Action,
				)
			}

			if transaction.Amount != tc.input.amount {
				t.Errorf(
					"expected transaction amount=%f, got amount=%f", tc.input.amount,
					transaction.Amount,
				)
			}
			if mockUser.AccountBalance != transaction.Amount {
				t.Errorf(
					"expected user account balance=%f, got account balance=%f", transaction.Amount,
					user.AnonymousUser.AccountBalance,
				)
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	mockUser := &user.User{
		ID:             1,
		Name:           "yusuf",
		Email:          "ym@gmail.com",
		AccountBalance: 100,
	}
	tests := []struct {
		name             string
		setupRepo        func(*MockRepo)
		setupUserService func(*MockUserService)
		input            struct {
			v           *validator.Validator
			userID      int64
			amount      float64
			performedBy string
		}
		expectedErr error
	}{
		{
			name:      "valid",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
		},
		{
			name:      "amount = 0",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 0, performedBy: "yusuf"},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "amount < 0",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: -100, performedBy: "yusuf"},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "amount > user account balance",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 200, performedBy: "yusuf"},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "GetUser failure",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserErr = user.ErrNoRecord
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
			expectedErr: user.ErrNoRecord,
		},
		{
			name: "Insert error",
			setupRepo: func(r *MockRepo) {
				r.InsertErr = errors.New("db Insert error")
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
			expectedErr: errors.New("db Insert error"),
		},
		{
			name:      "UpdateUser failure",
			setupRepo: func(r *MockRepo) {},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
				us.UpdateUserErr = errors.New("db UpdateUser error")
			},
			input: struct {
				v           *validator.Validator
				userID      int64
				amount      float64
				performedBy string
			}{v: validator.New(), userID: 1, amount: 100, performedBy: "yusuf"},
			expectedErr: errors.New("db UpdateUser error"),
		},
	}

	resetUser := func(u *user.User) {
		u.AccountBalance = 100
	}
	for _, tc := range tests {
		resetUser(mockUser)
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			userService := &MockUserService{}
			tc.setupRepo(repo)
			tc.setupUserService(userService)

			svc := Service{
				Repo:        repo,
				UserService: userService,
			}

			transaction, gotErr := svc.Withdraw(
				tc.input.v, tc.input.userID, tc.input.amount, tc.input.performedBy,
			)

			if tc.expectedErr != nil {
				if gotErr == nil || gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
			}

			withdrawAction := "WITHDRAW"
			if transaction.Action != withdrawAction {
				t.Errorf(
					"expected transaction action=%s, got action=%s", withdrawAction,
					transaction.Action,
				)
			}

			if transaction.Amount != tc.input.amount {
				t.Errorf(
					"expected transaction amount=%f, got amount=%f", tc.input.amount,
					transaction.Amount,
				)
			}
			if mockUser.AccountBalance != 0 {
				t.Errorf(
					"expected user account balance=%f, got account balance=%f", transaction.Amount,
					user.AnonymousUser.AccountBalance,
				)
			}
		})
	}
}
