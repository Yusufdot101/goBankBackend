package loanrequests

import (
	"errors"
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

// ---MOCKS---
type MockRepo struct {
	InsertErr error

	GetResult *LoanRequest
	GetErr    error

	UpdateTxResult *LoanRequest
	UpdateTxErr    error
}

func (r *MockRepo) Insert(loanRequest *LoanRequest) error {
	return r.InsertErr
}

func (r *MockRepo) Get(loanRequestID, userID int64) (*LoanRequest, error) {
	if r.GetErr != nil {
		return nil, r.GetErr
	}
	return r.GetResult, nil
}

func (r *MockRepo) UpdateTx(loanRequestID, userID int64, newStatus string) (*LoanRequest, error) {
	if r.UpdateTxErr != nil {
		return nil, r.UpdateTxErr
	}
	return r.UpdateTxResult, nil
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

type MockLoanService struct {
	GetLoanErr error
}

func (ls *MockLoanService) GetLoan(u *user.User, amount, dialyInterestRate float64) error {
	return ls.GetLoanErr
}

func TestNew(t *testing.T) {
	mockUser := &user.User{
		ID:             1,
		Name:           "yusuf",
		Email:          "ym@gmail.com",
		AccountBalance: 100,
	}

	tests := []struct {
		name      string
		setupRepo func(*MockRepo)
		input     struct {
			v                         *validator.Validator
			u                         *user.User
			amount, dialyInterestRate float64
		}
		expectedErr error
	}{
		{
			name:      "valid",
			setupRepo: func(r *MockRepo) {},
			input: struct {
				v                 *validator.Validator
				u                 *user.User
				amount            float64
				dialyInterestRate float64
			}{v: validator.New(), u: mockUser, amount: 100, dialyInterestRate: 5},
		},
		{
			name:      "amount = 0",
			setupRepo: func(r *MockRepo) {},
			input: struct {
				v                 *validator.Validator
				u                 *user.User
				amount            float64
				dialyInterestRate float64
			}{v: validator.New(), u: mockUser, amount: 0, dialyInterestRate: 5},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "negative amount",
			setupRepo: func(r *MockRepo) {},
			input: struct {
				v                 *validator.Validator
				u                 *user.User
				amount            float64
				dialyInterestRate float64
			}{v: validator.New(), u: mockUser, amount: -100, dialyInterestRate: 5},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name:      "negative dialy interest rate",
			setupRepo: func(r *MockRepo) {},
			input: struct {
				v                 *validator.Validator
				u                 *user.User
				amount            float64
				dialyInterestRate float64
			}{v: validator.New(), u: mockUser, amount: 100, dialyInterestRate: -5},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name: "insert failure",
			setupRepo: func(r *MockRepo) {
				r.InsertErr = errors.New("db error")
			},
			input: struct {
				v                 *validator.Validator
				u                 *user.User
				amount            float64
				dialyInterestRate float64
			}{v: validator.New(), u: mockUser, amount: 100, dialyInterestRate: 5},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tc.setupRepo(repo)
			svc := Service{Repo: repo}

			loanRequest, gotErr := svc.New(
				tc.input.v, tc.input.u, tc.input.amount, tc.input.dialyInterestRate,
			)
			if tc.expectedErr != nil {
				if gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error :%v", gotErr)
			}

			if loanRequest.UserID != mockUser.ID {
				t.Errorf("expected user id %d, got %d", mockUser.ID, loanRequest.UserID)
			}

			if loanRequest.Amount != tc.input.amount {
				t.Errorf("expected amount %f, got %f", tc.input.amount, loanRequest.Amount)
			}

			if loanRequest.DailyInterestRate != tc.input.dialyInterestRate {
				t.Errorf(
					"expected dialy interest rate %f, got %f", tc.input.dialyInterestRate,
					loanRequest.DailyInterestRate,
				)
			}

			defaultStatus := "PENDING"
			if loanRequest.Status != defaultStatus {
				t.Errorf("expected status %s, got %s", defaultStatus, loanRequest.Status)
			}
		})
	}
}

func TestAcceptLoanRequest(t *testing.T) {
	mockLoanRequest := &LoanRequest{
		ID:                1,
		UserID:            1,
		Amount:            100,
		DailyInterestRate: 5,
	}
	mockUser := &user.User{
		ID:             1,
		Name:           "yusuf",
		Email:          "ym@gmail",
		AccountBalance: 0,
	}

	tests := []struct {
		name             string
		setupRepo        func(*MockRepo)
		setupUserService func(*MockUserService)
		setupLoanService func(*MockLoanService)
		input            struct {
			loanRequestID, userID int64
		}
		loanRequestOriginalStatus string
		expectedErr               error
	}{
		{
			name: "valid",
			setupRepo: func(r *MockRepo) {
				r.GetResult = mockLoanRequest
				r.UpdateTxResult = &LoanRequest{
					ID: mockLoanRequest.ID, UserID: mockLoanRequest.UserID,
					Amount: mockLoanRequest.Amount, Status: "ACCEPTED",
					DailyInterestRate: mockLoanRequest.DailyInterestRate,
				}
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			setupLoanService: func(ls *MockLoanService) {},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			loanRequestOriginalStatus: "PENDING",
		},
		{
			name: "loan request already responded to",
			setupRepo: func(r *MockRepo) {
				r.GetResult = mockLoanRequest
				r.UpdateTxResult = &LoanRequest{
					ID: mockLoanRequest.ID, UserID: mockLoanRequest.UserID,
					Amount: mockLoanRequest.Amount, Status: "ACCEPTED",
					DailyInterestRate: mockLoanRequest.DailyInterestRate,
				}
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			setupLoanService: func(ls *MockLoanService) {},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			loanRequestOriginalStatus: "ACCEPTED",
			expectedErr:               user.ErrNoRecord,
		},
		{
			name: "Get loan failure",
			setupRepo: func(r *MockRepo) {
				r.GetErr = errors.New("db error")
			},
			setupUserService: func(us *MockUserService) {},
			setupLoanService: func(ls *MockLoanService) {},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			expectedErr: errors.New("db error"),
		},
		{
			name: "Get user failure",
			setupRepo: func(r *MockRepo) {
				r.GetResult = mockLoanRequest
				r.UpdateTxResult = &LoanRequest{
					ID: mockLoanRequest.ID, UserID: mockLoanRequest.UserID,
					Amount: mockLoanRequest.Amount, Status: "ACCEPTED",
					DailyInterestRate: mockLoanRequest.DailyInterestRate,
				}
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserErr = user.ErrNoRecord
			},
			setupLoanService: func(ls *MockLoanService) {},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			expectedErr:               user.ErrNoRecord,
			loanRequestOriginalStatus: "PENDING",
		},
		{
			name: "update user failure",
			setupRepo: func(r *MockRepo) {
				r.GetResult = mockLoanRequest
				r.UpdateTxResult = &LoanRequest{
					ID: mockLoanRequest.ID, UserID: mockLoanRequest.UserID,
					Amount: mockLoanRequest.Amount, Status: "ACCEPTED",
					DailyInterestRate: mockLoanRequest.DailyInterestRate,
				}
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
				us.UpdateUserErr = errors.New("db error")
			},
			setupLoanService: func(ls *MockLoanService) {},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			expectedErr:               errors.New("db error"),
			loanRequestOriginalStatus: "PENDING",
		},
		{
			name: "update loan failure",
			setupRepo: func(r *MockRepo) {
				r.GetResult = mockLoanRequest
				r.UpdateTxErr = errors.New("db error")
			},
			setupUserService: func(us *MockUserService) {},
			setupLoanService: func(ls *MockLoanService) {},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			loanRequestOriginalStatus: "PENDING",
			expectedErr:               errors.New("db error"),
		},
		{
			name: "GetLoan failure",
			setupRepo: func(r *MockRepo) {
				r.GetResult = mockLoanRequest
				r.UpdateTxResult = &LoanRequest{
					ID: mockLoanRequest.ID, UserID: mockLoanRequest.UserID,
					Amount: mockLoanRequest.Amount, Status: "ACCEPTED",
					DailyInterestRate: mockLoanRequest.DailyInterestRate,
				}
			},
			setupUserService: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			setupLoanService: func(ls *MockLoanService) {
				ls.GetLoanErr = errors.New("db error")
			},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: mockLoanRequest.ID, userID: mockUser.ID},
			loanRequestOriginalStatus: "PENDING",
			expectedErr:               errors.New("db error"),
		},
	}

	for _, tc := range tests {
		mockLoanRequest.Status = tc.loanRequestOriginalStatus
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			userSvc := &MockUserService{}
			loanSvc := &MockLoanService{}
			tc.setupRepo(repo)
			tc.setupUserService(userSvc)
			tc.setupLoanService(loanSvc)

			svc := Service{
				Repo:        repo,
				UserService: userSvc,
				LoanService: loanSvc,
			}

			loanRequest, gotErr := svc.AcceptLoanRequest(tc.input.loanRequestID, tc.input.userID)

			if tc.expectedErr != nil {
				if gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error :%v", gotErr)
			}

			if loanRequest.UserID != mockUser.ID {
				t.Errorf("expected user id %d, got %d", mockUser.ID, loanRequest.UserID)
			}

			if loanRequest.Amount != mockLoanRequest.Amount {
				t.Errorf("expected amount %f, got %f", mockLoanRequest.Amount, loanRequest.Amount)
			}

			if loanRequest.DailyInterestRate != mockLoanRequest.DailyInterestRate {
				t.Errorf(
					"expected dialy interest rate %f, got %f", mockLoanRequest.DailyInterestRate,
					loanRequest.DailyInterestRate,
				)
			}

			if loanRequest.Status != "ACCEPTED" {
				t.Errorf("expected status %s, got %s", "ACCEPTED", loanRequest.Status)
			}

			// check if the money is getting added to the users account
			if mockUser.AccountBalance != loanRequest.Amount {
				t.Errorf(
					"expected user account balance %f, got %f",
					loanRequest.Amount, mockUser.AccountBalance,
				)
			}
		})
	}
}

func TestDeclineLoanRequest(t *testing.T) {
	mockLoanRequest := &LoanRequest{
		ID:                1,
		UserID:            1,
		Amount:            100,
		DailyInterestRate: 5,
		Status:            "PENDING",
	}

	tests := []struct {
		name        string
		setupRepo   func(*MockRepo)
		input       struct{ loanRequestID, userID int64 }
		expectedErr error
	}{
		{
			name: "valid",
			setupRepo: func(r *MockRepo) {
				r.UpdateTxResult = &LoanRequest{
					ID: mockLoanRequest.ID, UserID: mockLoanRequest.UserID,
					Status: "DECLINED",
				}
			},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: 1, userID: 1},
		},
		{
			name: "update failure",
			setupRepo: func(r *MockRepo) {
				r.UpdateTxErr = errors.New("db error")
			},
			input: struct {
				loanRequestID int64
				userID        int64
			}{loanRequestID: 1, userID: 1},
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
			loanRequest, gotErr := svc.DeclineLoanRequest(tc.input.loanRequestID, tc.input.userID)
			if tc.expectedErr != nil {
				if gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error :%v", gotErr)
			}

			if loanRequest.ID != mockLoanRequest.ID {
				t.Errorf("expected loan id %d, got %d", mockLoanRequest.ID, loanRequest.ID)
			}

			if loanRequest.UserID != mockLoanRequest.UserID {
				t.Errorf("expected user id %d, got %d", mockLoanRequest.ID, loanRequest.ID)
			}

			if loanRequest.Status != "DECLINED" {
				t.Errorf("expected status %s, got %s", "ACCEPTED", loanRequest.Status)
			}
		})
	}
}
