package loan

import (
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type MockRepo struct {
	InsertCalled bool
	InsertErr    error

	InsertDeletionCalled bool
	InsertDeletionErr    error

	GetByIDCalled bool
	GetByIDResult *Loan
	GetByIDErr    error

	DeleteLoanCalled bool
	DeleteLoanErr    error

	MakePaymentTxCalled bool
	MakePaymentTxResult *Loan
	MakePaymentTxErr    error
}

func (m *MockRepo) Insert(loan *Loan) error {
	m.InsertCalled = true
	return m.InsertErr
}

func (m *MockRepo) InsertDeletion(loan *LoanDeletion) error {
	m.InsertDeletionCalled = true
	return m.InsertDeletionErr
}

func (m *MockRepo) GetByID(loanID, userID int64) (*Loan, error) {
	m.GetByIDCalled = true
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	return m.GetByIDResult, nil
}

func (m *MockRepo) DeleteLoan(loanID, debtorID int64) error {
	m.DeleteLoanCalled = true
	return m.DeleteLoanErr
}

func (m *MockRepo) MakePaymentTx(loanID, userID int64, payment, totalOwed float64) (*Loan, error) {
	m.MakePaymentTxCalled = true
	if m.MakePaymentTxErr != nil {
		return nil, m.MakePaymentTxErr
	}

	return m.MakePaymentTxResult, nil
}

type MockUserService struct {
	GetUserCalled bool
	GetUserResult *user.User
	GetUserErr    error

	UpdateUserCalled bool
	UpdateUserResult *user.User
	UpdateUserErr    error
}

func (us *MockUserService) GetUser(userID int64) (*user.User, error) {
	us.GetUserCalled = true
	if us.GetUserErr != nil {
		return nil, us.GetUserErr
	}

	return us.GetUserResult, nil
}

func (us *MockUserService) UpdateUser(
	userID int64, userName, userEmail string, userPasswordHash []byte,
	userAccountBalance float64, userActivated bool,
) (*user.User, error) {
	us.UpdateUserCalled = true
	if us.UpdateUserErr != nil {
		return nil, us.UpdateUserErr
	}

	return us.UpdateUserResult, nil
}

func TestMakepayment(t *testing.T) {
	mockLoan := &Loan{
		ID:              1,
		UserID:          1,
		Amount:          200,
		Action:          "took",
		RemainingAmount: 200,
	}
	mockUser := &user.User{
		ID:             1,
		Name:           "yusuf",
		Email:          "ym@gmail.com",
		AccountBalance: 100,
	}

	tests := []struct {
		name         string
		setupRepo    func(*MockRepo)
		setupUserSvc func(*MockUserService)
		input        struct {
			v              *validator.Validator
			loanID, userID int64
			payment        float64
		}
		finalLoanRemainingAmount float64
		expectedErr              error
	}{
		{
			name: "vaild input",
			setupRepo: func(r *MockRepo) {
				r.GetByIDResult = mockLoan
				r.MakePaymentTxResult = &Loan{RemainingAmount: 150}
			},
			setupUserSvc: func(us *MockUserService) {
				us.GetUserResult = mockUser
				us.UpdateUserResult = &user.User{AccountBalance: 50}
			},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 50},
			finalLoanRemainingAmount: 150,
		},
		{
			name: "insufficient funds",
			setupRepo: func(r *MockRepo) {
				r.GetByIDResult = mockLoan
			},
			setupUserSvc: func(us *MockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 200},
			finalLoanRemainingAmount: 200,
			expectedErr:              validator.ErrFailedValidation,
		},
		{
			name:         "negative amount",
			setupRepo:    func(r *MockRepo) {},
			setupUserSvc: func(us *MockUserService) {},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: -100},
			finalLoanRemainingAmount: 200,
			expectedErr:              validator.ErrFailedValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// reset the user AccountBalance to avoid confusion and unexpected behaviour
			mockUser.AccountBalance = 100
			repo := &MockRepo{}
			userSvc := &MockUserService{}
			tc.setupRepo(repo)
			tc.setupUserSvc(userSvc)

			svc := Service{
				Repo:        repo,
				UserService: userSvc,
			}

			gotLoan, gotErr := svc.MakePayment(
				tc.input.v, tc.input.loanID, tc.input.userID, tc.input.payment,
			)
			if gotErr != tc.expectedErr {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
			} else if gotErr != nil {
				return
			}

			if gotLoan.RemainingAmount != tc.finalLoanRemainingAmount {
				t.Fatalf(
					"expected remaining amount %f, got %f", tc.finalLoanRemainingAmount,
					gotLoan.RemainingAmount,
				)
			}
		})
	}
}

func TestDeleteLoan(t *testing.T) {
	mockLoan := &Loan{
		ID:              1,
		UserID:          1,
		Amount:          200,
		Action:          "took",
		RemainingAmount: 200,
	}

	tests := []struct {
		name      string
		setupRepo func(*MockRepo)
		input     struct {
			v                             *validator.Validator
			loanID, debtorID, deletedByID int64
			reason                        string
		}
		expectedErr error
	}{
		{
			name: "valid",
			setupRepo: func(r *MockRepo) {
				r.GetByIDResult = mockLoan
			},
			input: struct {
				v           *validator.Validator
				loanID      int64
				debtorID    int64
				deletedByID int64
				reason      string
			}{v: validator.New(), loanID: 1, debtorID: 1, deletedByID: 1, reason: "some reason"},
		},
		{
			name: "load with id not found",
			setupRepo: func(r *MockRepo) {
				r.GetByIDErr = user.ErrNoRecord
			},
			input: struct {
				v           *validator.Validator
				loanID      int64
				debtorID    int64
				deletedByID int64
				reason      string
			}{v: validator.New(), loanID: 2, debtorID: 1, deletedByID: 1, reason: "some reason"},
			expectedErr: user.ErrNoRecord,
		},
		{
			name:      "reason not given",
			setupRepo: func(r *MockRepo) {},
			input: struct {
				v           *validator.Validator
				loanID      int64
				debtorID    int64
				deletedByID int64
				reason      string
			}{v: validator.New(), loanID: 2, debtorID: 1, deletedByID: 1, reason: ""},
			expectedErr: validator.ErrFailedValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepo{}
			tc.setupRepo(repo)
			svc := Service{Repo: repo}

			gotLoan, gotErr := svc.DeleteLoan(
				tc.input.v, tc.input.loanID, tc.input.debtorID, tc.input.deletedByID,
				tc.input.reason,
			)
			if gotErr != tc.expectedErr {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
			} else if gotErr != nil {
				return
			}

			if gotLoan.Amount != mockLoan.Amount {
				t.Errorf("expected amount %f, got %f", mockLoan.Amount, gotLoan.Amount)
			}

			if gotLoan.LoanID != mockLoan.ID {
				t.Errorf("expected loan ID %d, got %d", mockLoan.ID, gotLoan.LoanID)
			}
		})
	}
}
