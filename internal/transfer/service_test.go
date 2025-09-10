package transfer

import (
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

// ---MOCKS---

type MockRepo struct {
	InsertErr error
}

func (r *MockRepo) Insert(transfer *Transfer) error {
	return r.InsertErr
}

type MockUserService struct {
	GetUserByEmailResult *user.User
	GetUserByEmailErr    error

	TransferMoneyResult *user.User
	TransferMoneyErr    error
}

func (us *MockUserService) GetUserByEmail(email string) (*user.User, error) {
	if us.GetUserByEmailErr != nil {
		return nil, us.GetUserByEmailErr
	}
	return us.GetUserByEmailResult, nil
}

func (us *MockUserService) TransferMoney(
	fromUser, toUser *user.User, amount float64,
) (*user.User, error) {
	if us.TransferMoneyErr != nil {
		return nil, us.TransferMoneyErr
	}

	toUser.AccountBalance += amount
	return us.TransferMoneyResult, us.TransferMoneyErr
}

func TestTransferMoney(t *testing.T) {
	fromUser := &user.User{
		ID: 1, Name: "yusuf", Email: "a@b.com", AccountBalance: 100,
	}
	toUser := &user.User{
		ID: 2, Name: "mohamed", Email: "b@a.com", AccountBalance: 50,
	}

	tests := []struct {
		name         string
		setupRepo    func(*MockRepo)
		setupUserSvc func(*MockUserService)
		input        struct {
			v           *validator.Validator
			fromUser    *user.User
			toUserEmail string
			amount      float64
		}
		finalFrom   float64
		finalTo     float64
		expectedErr error
	}{
		{
			name:      "valid input",
			setupRepo: func(m *MockRepo) {},
			setupUserSvc: func(us *MockUserService) {
				us.GetUserByEmailResult = toUser
				us.TransferMoneyResult = &user.User{AccountBalance: 90}
			},
			input: struct {
				v           *validator.Validator
				fromUser    *user.User
				toUserEmail string
				amount      float64
			}{v: validator.New(), fromUser: fromUser, toUserEmail: toUser.Email, amount: 10},
			finalFrom:   90,
			finalTo:     60,
			expectedErr: nil,
		},
		{
			name:      "insuffient funds",
			setupRepo: func(m *MockRepo) {},
			setupUserSvc: func(us *MockUserService) {
				us.GetUserByEmailResult = toUser
			},
			input: struct {
				v           *validator.Validator
				fromUser    *user.User
				toUserEmail string
				amount      float64
			}{v: validator.New(), fromUser: fromUser, toUserEmail: toUser.Email, amount: 1000},
			finalFrom:   100,
			finalTo:     50,
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "to user not found",
			setupRepo: func(m *MockRepo) {},
			setupUserSvc: func(us *MockUserService) {
				us.GetUserByEmailErr = user.ErrNoRecord
			},
			input: struct {
				v           *validator.Validator
				fromUser    *user.User
				toUserEmail string
				amount      float64
			}{v: validator.New(), fromUser: fromUser, toUserEmail: "random@email.gmail", amount: 10},
			finalFrom:   100,
			expectedErr: user.ErrNoRecord,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			userSvc := &MockUserService{}
			tc.setupRepo(repo)
			tc.setupUserSvc(userSvc)
			svc := Service{
				Repo:        repo,
				UserService: userSvc,
			}

			_, gotUser, gotErr := svc.TransferMoney(
				tc.input.v, tc.input.fromUser, tc.input.toUserEmail, tc.input.amount,
			)

			if gotErr != tc.expectedErr {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
			}
			if gotErr != nil {
				return
			}

			if toUser.AccountBalance != tc.finalTo {
				t.Errorf(
					"expected balances from=%v, to=%v; got from=%v, to=%v", tc.finalFrom, tc.finalTo,
					gotUser.AccountBalance, toUser.AccountBalance,
				)
			}
		})
	}
}
