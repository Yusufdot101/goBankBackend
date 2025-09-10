package loan

import (
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func TestValidateLoan(t *testing.T) {
	mockLoan := &Loan{
		Action:            "took",
		DailyInterestRate: 5,
		Amount:            100,
	}

	tests := []struct {
		name      string
		setupLoan func(*Loan)
		input     struct {
			v    *validator.Validator
			loan *Loan
		}
		wantValid      bool
		expectedErrMsg map[string]string
	}{
		{
			name:      "valid",
			setupLoan: func(l *Loan) {},
			wantValid: true,
		},
		{
			name:      "amount = 0",
			setupLoan: func(l *Loan) { mockLoan.Amount = 0 },
			wantValid: false,
			expectedErrMsg: map[string]string{
				"amount": "must be given",
			},
		},
		{
			name:      "amount < 0",
			setupLoan: func(l *Loan) { mockLoan.Amount = -100 },
			wantValid: false,
			expectedErrMsg: map[string]string{
				"amount": "must be more than 0",
			},
		},
		{
			name:      "DailyInterestRate < 0",
			setupLoan: func(l *Loan) { mockLoan.DailyInterestRate = -10 },
			wantValid: false,
			expectedErrMsg: map[string]string{
				"daily interest rate": "must be more than 0",
			},
		},
		{
			name:      "unrecognised action",
			setupLoan: func(l *Loan) { mockLoan.Action = "random action" },
			wantValid: false,
			expectedErrMsg: map[string]string{
				"action": "invalid",
			},
		},
	}

	resetLoan := func(loan *Loan) {
		loan.Action = "took"
		loan.DailyInterestRate = 5
		loan.Amount = 100
	}

	for _, tc := range tests {
		resetLoan(mockLoan)
		t.Run(tc.name, func(t *testing.T) {
			tc.setupLoan(mockLoan)
			v := validator.New()
			ValidateLoan(v, mockLoan)
			if v.IsValid() != tc.wantValid {
				t.Fatalf("expected valid=%v, got valid=%v", tc.wantValid, v.IsValid())
			} else if v.IsValid() {
				return
			}

			for key, val := range tc.expectedErrMsg {
				if v.Errors[key] != val {
					t.Errorf(
						"expected message=%s for key=%v, got message=%s", val, key, v.Errors[key],
					)
				}
			}
		})
	}
}

func TestValidateLoanDeletion(t *testing.T) {
	mockLoanDeletion := &LoanDeletion{
		LoanID:      1,
		DebtorID:    1,
		DeletedByID: 1,
		Reason:      "some reason",
	}

	tests := []struct {
		name              string
		setupLoanDeletion func(*LoanDeletion)
		wantValid         bool
		expectedErrMsg    map[string]string
	}{
		{
			name:              "valid",
			setupLoanDeletion: func(ld *LoanDeletion) {},
			wantValid:         true,
		},
		{
			name: "reason not given",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.Reason = ""
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"reason": "must be given",
			},
		},
		{
			name: "loan id = 0",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.LoanID = 0
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"loan ID": "must be given",
			},
		},
		{
			name: "debtor id = 0",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.DebtorID = 0
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"debtor ID": "must be given",
			},
		},
		{
			name: "deleted by  id = 0",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.DeletedByID = 0
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"deleted by ID": "must be given",
			},
		},
		{
			name: "loan id < 0",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.LoanID = -1
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"loan ID": "must be more than 0",
			},
		},
		{
			name: "debtor id < 0",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.DebtorID = -1
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"debtor ID": "must be more than 0",
			},
		},
		{
			name: "deleted by id < 0",
			setupLoanDeletion: func(ld *LoanDeletion) {
				mockLoanDeletion.DeletedByID = -1
			},
			wantValid: false,
			expectedErrMsg: map[string]string{
				"deleted by ID": "must be more than 0",
			},
		},
	}

	resetLoanDeletion := func(loanDeletion *LoanDeletion) {
		loanDeletion.LoanID = 1
		loanDeletion.DebtorID = 1
		loanDeletion.DeletedByID = 1
		loanDeletion.Reason = "some reason"
	}

	for _, tc := range tests {
		resetLoanDeletion(mockLoanDeletion)
		t.Run(tc.name, func(t *testing.T) {
			tc.setupLoanDeletion(mockLoanDeletion)
			v := validator.New()
			ValidateLoanDeletion(v, mockLoanDeletion)
			if v.IsValid() != tc.wantValid {
				t.Fatalf("expected valid=%v, got valid=%v", tc.wantValid, v.IsValid())
			} else if v.IsValid() {
				return
			}

			for key, val := range tc.expectedErrMsg {
				if v.Errors[key] != val {
					t.Errorf(
						"expected message=%s for key=%v, got message=%s", val, key, v.Errors[key],
					)
				}
			}
		})
	}
}
