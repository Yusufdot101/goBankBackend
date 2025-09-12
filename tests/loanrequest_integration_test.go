package tests

import (
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/loan"
	"github.com/Yusufdot101/goBankBackend/internal/loanrequests"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
	_ "github.com/lib/pq"
)

func TestLoanRequest(t *testing.T) {
	loanrequestRepo = &loanrequests.Repository{
		DB: testDB,
	}
	loanSvc = &loan.Service{
		Repo: &loan.Repository{DB: loanrequestRepo.DB},
	}
	userSvc = &user.Service{
		Repo: &user.Repository{DB: loanrequestRepo.DB},
	}
	loanrequestSvc = &loanrequests.Service{
		Repo:        loanrequestRepo,
		LoanService: loanSvc,
		UserService: userSvc,
	}

	user1 = &user.User{
		ID:    1,
		Name:  "yusuf",
		Email: "y@gmail.com",
	}

	seedUsersTable := func(us *user.Service) {
		// seed the users table, this will be used in transferring of money
		us.Repo.Insert(user1)
	}

	tests := []struct {
		name            string
		setup           func()
		setupUserSevice func(*user.Service)
		input           struct {
			u                         *user.User
			amount, dailyInterestRate float64
		}
		expectedErr error
	}{
		{
			name: "valid",
			setup: func() {
				resetDB() // reset the db and start on clean slate
			},
			setupUserSevice: func(us *user.Service) {
				seedUsersTable(us)
			},
			input: struct {
				u                 *user.User
				amount            float64
				dailyInterestRate float64
			}{
				u:                 user1,
				amount:            100,
				dailyInterestRate: 5,
			},
		},
		{
			name: "amount = 0",
			setup: func() {
				resetDB() // reset the db and start on clean slate
			},
			setupUserSevice: func(us *user.Service) {
				seedUsersTable(us)
			},
			input: struct {
				u                 *user.User
				amount            float64
				dailyInterestRate float64
			}{
				u:                 user1,
				amount:            0,
				dailyInterestRate: 5,
			},
			expectedErr: validator.ErrFailedValidation,
		},
		{
			name: "amount < 0",
			setup: func() {
				resetDB() // reset the db and start on clean slate
			},
			setupUserSevice: func(us *user.Service) {
				seedUsersTable(us)
			},
			input: struct {
				u                 *user.User
				amount            float64
				dailyInterestRate float64
			}{
				u:                 user1,
				amount:            -100,
				dailyInterestRate: 5,
			},
			expectedErr: validator.ErrFailedValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			tc.setupUserSevice(userSvc)
			v := validator.New()
			// step 1: loan creation
			loanRequest, gotErr := loanrequestSvc.New(
				v, tc.input.u, tc.input.amount, tc.input.dailyInterestRate,
			)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}

			// fetch the loan
			loanRequest, gotErr = loanrequestSvc.Repo.Get(loanRequest.ID, tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}
			passed := checkLoanRequest(
				t, loanRequest, tc.input.u.ID, tc.input.amount, tc.input.dailyInterestRate,
				"PENDING", "Get",
			)
			if !passed {
				return
			}

			// step 2: accept the loan
			loanRequest, gotErr = loanrequestSvc.AcceptLoanRequest(loanRequest.ID, tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}

			// fetch the loan
			loanRequest, gotErr = loanrequestSvc.Repo.Get(loanRequest.ID, tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}
			passed = checkLoanRequest(
				t, loanRequest, tc.input.u.ID, tc.input.amount, tc.input.dailyInterestRate,
				"ACCEPTED", "Get",
			)
			if !passed {
				return
			}

			// check if the loan was added to the loans table
			gotLoan, gotErr := loanSvc.Repo.GetByID(1, tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}
			passed = checkLoan(
				t, gotLoan, tc.input.u.ID, tc.input.amount, tc.input.dailyInterestRate, "GetByID",
			)
			if !passed {
				return
			}

			// check if the money was added to the users account balance
			gotUser, gotErr := userSvc.GetUser(tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}
			if gotUser.AccountBalance != tc.input.amount {
				t.Errorf(
					"expected user account balance=%f, got account balance=%f",
					tc.input.amount, gotUser.AccountBalance,
				)
			}

			// step 3: new loan request
			loanRequest, gotErr = loanrequestSvc.New(
				v, tc.input.u, tc.input.amount, tc.input.dailyInterestRate,
			)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}

			// decline it
			loanRequest, gotErr = loanrequestSvc.DeclineLoanRequest(loanRequest.ID, tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}

			// fetch it
			loanRequest, gotErr = loanrequestSvc.Repo.Get(loanRequest.ID, tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New") {
				return
			}
			passed = checkLoanRequest(
				t, loanRequest, tc.input.u.ID, tc.input.amount, tc.input.dailyInterestRate,
				"DECLINED", "Get",
			)
			if !passed {
				return
			}

			// make sure its not added to the loans table
			gotLoan, gotErr = loanSvc.Repo.GetByID(1, tc.input.u.ID)
			if !checkErr(t, gotErr, user.ErrNoRecord, "New") {
				return
			}
			passed = checkLoan(
				t, gotLoan, tc.input.u.ID, tc.input.amount, tc.input.dailyInterestRate, "GetByID 2",
			)
			if !passed {
				return
			}

			// make sure the money was not added to the users account balance
			gotUser, gotErr = userSvc.GetUser(tc.input.u.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "New 2") {
				return
			}
			// it should only have the balance from the first loan
			if gotUser.AccountBalance > tc.input.amount {
				t.Errorf(
					"expected user account balance=%f, got account balance=%f",
					tc.input.amount, gotUser.AccountBalance,
				)
			}

			// if we expected an error but no error occured
			if tc.expectedErr != nil {
				t.Fatalf("expected error %v, got nil", tc.expectedErr)
			}
		})
	}
}

func checkLoan(
	t *testing.T, gotLoan *loan.Loan, userID int64, amount, dailyInterestRate float64, msg string,
) bool {
	passed := true
	if gotLoan.UserID != userID {
		t.Errorf("%s: expected user id=%d, got id=%d", msg, userID, gotLoan.UserID)
		passed = false
	}
	if gotLoan.Amount != amount {
		t.Errorf(
			"%s: expected account balance=%f, got account balance=%f", msg, amount,
			gotLoan.Amount,
		)
		passed = false
	}
	if gotLoan.DailyInterestRate != dailyInterestRate {
		t.Errorf(
			"%s: expected daily interest rate=%f, got daily interest rate=%f",
			msg, dailyInterestRate, gotLoan.DailyInterestRate,
		)
		passed = false
	}
	return passed
}

func checkLoanRequest(
	t *testing.T, gotLoanRequest *loanrequests.LoanRequest, userID int64, amount, dailyInterestRate float64,
	status, msg string,
) bool {
	passed := true
	if gotLoanRequest.UserID != userID {
		t.Errorf("%s: expected user id=%d, got id=%d", msg, userID, gotLoanRequest.UserID)
		passed = false
	}
	if gotLoanRequest.Amount != amount {
		t.Errorf(
			"%s: expected account balance=%f, got account balance=%f", msg, amount,
			gotLoanRequest.Amount,
		)
		passed = false
	}
	if gotLoanRequest.DailyInterestRate != dailyInterestRate {
		t.Errorf(
			"%s: expected daily interest rate=%f, got daily interest rate=%f",
			msg, dailyInterestRate, gotLoanRequest.DailyInterestRate,
		)
		passed = false
	}
	if gotLoanRequest.Status != status {
		t.Errorf(
			"%s: expected loan status=%s, got account status=%s", msg, status, gotLoanRequest.Status,
		)
		passed = false
	}
	return passed
}
