package loan

import (
	"errors"
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type mockRepo struct {
	InsertErr error

	InsertDeletionErr error

	GetByIDResult *Loan
	GetByIDErr    error

	DeleteLoanErr error

	MakePaymentTxResult *Loan
	MakePaymentTxErr    error
}

func (m *mockRepo) Insert(loan *Loan) error {
	return m.InsertErr
}

func (m *mockRepo) InsertDeletion(loan *LoanDeletion) error {
	return m.InsertDeletionErr
}

func (m *mockRepo) GetByID(loanID, userID int64) (*Loan, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	return m.GetByIDResult, nil
}

func (m *mockRepo) DeleteLoan(loanID, debtorID int64) error {
	return m.DeleteLoanErr
}

func (m *mockRepo) MakePaymentTx(loanID, userID int64, payment, totalOwed float64) (*Loan, error) {
	if m.MakePaymentTxErr != nil {
		return nil, m.MakePaymentTxErr
	}

	return m.MakePaymentTxResult, nil
}

type mockUserService struct {
	GetUserResult *user.User
	GetUserErr    error

	UpdateUserErr error
}

func (us *mockUserService) GetUser(userID int64) (*user.User, error) {
	if us.GetUserErr != nil {
		return nil, us.GetUserErr
	}

	return us.GetUserResult, nil
}

func (us *mockUserService) UpdateUser(
	userID int64, userName, userEmail string, userPasswordHash []byte,
	userAccountBalance float64, userActivated bool,
) (*user.User, error) {
	if us.UpdateUserErr != nil {
		return nil, us.UpdateUserErr
	}
	u, err := us.GetUser(userID)
	if err != nil {
		return nil, err
	}
	u.AccountBalance = userAccountBalance
	u.Name = userName
	u.Email = userEmail
	u.Password.Hash = userPasswordHash
	u.Activated = userActivated
	return u, nil
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
		setupRepo    func(*mockRepo)
		setupUserSvc func(*mockUserService)
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
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
				r.MakePaymentTxResult = &Loan{RemainingAmount: 150}
			},
			setupUserSvc: func(us *mockUserService) {
				us.GetUserResult = mockUser
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
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
			},
			setupUserSvc: func(us *mockUserService) {
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
			name: "loan already paid off",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = &Loan{RemainingAmount: 0}
			},
			setupUserSvc: func(us *mockUserService) {
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
			setupRepo:    func(r *mockRepo) {},
			setupUserSvc: func(us *mockUserService) {},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: -100},
			finalLoanRemainingAmount: 200,
			expectedErr:              validator.ErrFailedValidation,
		},
		{
			name: "GetByID failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDErr = user.ErrNoRecord
			},
			setupUserSvc: func(us *mockUserService) {},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 100},
			finalLoanRemainingAmount: 200,
			expectedErr:              user.ErrNoRecord,
		},
		{
			name: "GetUser failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
			},
			setupUserSvc: func(us *mockUserService) {
				us.GetUserErr = user.ErrNoRecord
			},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 100},
			finalLoanRemainingAmount: 200,
			expectedErr:              user.ErrNoRecord,
		},
		{
			name: "MakePaymentTx failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
				r.MakePaymentTxErr = errors.New("db MakePaymentTx error")
			},
			setupUserSvc: func(us *mockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 100},
			finalLoanRemainingAmount: 200,
			expectedErr:              errors.New("db MakePaymentTx error"),
		},
		{
			name: "Insert failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
				r.MakePaymentTxResult = mockLoan
				r.InsertErr = errors.New("db Insert error")
			},
			setupUserSvc: func(us *mockUserService) {
				us.GetUserResult = mockUser
			},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 100},
			finalLoanRemainingAmount: 200,
			expectedErr:              errors.New("db Insert error"),
		},
		{
			name: "UpdateUser failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
				r.MakePaymentTxResult = mockLoan
			},
			setupUserSvc: func(us *mockUserService) {
				us.GetUserResult = mockUser
				us.UpdateUserErr = errors.New("db UpdateUser error")
			},
			input: struct {
				v       *validator.Validator
				loanID  int64
				userID  int64
				payment float64
			}{v: validator.New(), loanID: 1, userID: 1, payment: 100},
			finalLoanRemainingAmount: 200,
			expectedErr:              errors.New("db UpdateUser error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// reset the user AccountBalance to avoid confusion and unexpected behaviour
			mockUser.AccountBalance = 100
			repo := &mockRepo{}
			userSvc := &mockUserService{}
			tc.setupRepo(repo)
			tc.setupUserSvc(userSvc)

			svc := Service{
				Repo:        repo,
				UserService: userSvc,
			}

			gotLoan, gotErr := svc.MakePayment(
				tc.input.v, tc.input.loanID, tc.input.userID, tc.input.payment,
			)
			if tc.expectedErr != nil {
				if gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil {
				t.Fatalf("unexpected error %v", gotErr)
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
		setupRepo func(*mockRepo)
		input     struct {
			v                             *validator.Validator
			loanID, debtorID, deletedByID int64
			reason                        string
		}
		expectedErr error
	}{
		{
			name: "valid",
			setupRepo: func(r *mockRepo) {
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
			setupRepo: func(r *mockRepo) {
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
			setupRepo: func(r *mockRepo) {},
			input: struct {
				v           *validator.Validator
				loanID      int64
				debtorID    int64
				deletedByID int64
				reason      string
			}{v: validator.New(), loanID: 2, debtorID: 1, deletedByID: 1, reason: ""},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name: "InsertDeletion failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
				r.InsertDeletionErr = errors.New("db InsertDeletion error")
			},
			input: struct {
				v           *validator.Validator
				loanID      int64
				debtorID    int64
				deletedByID int64
				reason      string
			}{v: validator.New(), loanID: 1, debtorID: 1, deletedByID: 1, reason: "some reason"},
			expectedErr: errors.New("db InsertDeletion error"),
		},
		{
			name: "DeleteLoan failure",
			setupRepo: func(r *mockRepo) {
				r.GetByIDResult = mockLoan
				r.DeleteLoanErr = errors.New("db DeleteLoan error")
			},
			input: struct {
				v           *validator.Validator
				loanID      int64
				debtorID    int64
				deletedByID int64
				reason      string
			}{v: validator.New(), loanID: 1, debtorID: 1, deletedByID: 1, reason: "some reason"},
			expectedErr: errors.New("db DeleteLoan error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRepo{}
			tc.setupRepo(repo)
			svc := Service{Repo: repo}

			gotLoan, gotErr := svc.DeleteLoan(
				tc.input.v, tc.input.loanID, tc.input.debtorID, tc.input.deletedByID,
				tc.input.reason,
			)

			if tc.expectedErr != nil {
				if gotErr.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, gotErr)
				}
				return
			} else if gotErr != nil { // we expected no error but got error
				t.Fatalf("unexpected error %v", gotErr)
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
