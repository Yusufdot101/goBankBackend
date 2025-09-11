package loan

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
	_ "github.com/lib/pq"
)

var (
	testDB *sql.DB
	repo   *Repository
	svc    *Service
	user1  *user.User
	user2  *user.User
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("GOBANK_TEST_DB")
	if dsn == "" {
		log.Fatal("GOBANK_TEST_DB not set")
	}

	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to test DB: %v", err)
	}

	repo = &Repository{
		DB: testDB,
	}

	resetDB()

	code := m.Run()
	resetDB()
	os.Exit(code)
}

func resetDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	testDB.ExecContext(ctx, `TRUNCATE loans, deleted_loans, users RESTART IDENTITY CASCADE`)
}

// actual integration tests

func TestLoanLifecycle(t *testing.T) {
	user1 = &user.User{
		ID:             1,
		Name:           "yusuf",
		Email:          "y@gmail.com",
		AccountBalance: 100, // needed to make the payment in the test
	}
	user1.Password.Set("12345678", 12)

	user2 = &user.User{
		ID:    2,
		Name:  "mohamed",
		Email: "m@gmail.com",
	}
	user2.Password.Set("12345678", 12)

	setupUserSevice := func(us *user.Service) {
		// seed the users table
		us.Repo.Insert(user1)
		us.Repo.Insert(user2)
	}

	tests := []struct {
		name  string
		input struct {
			user                               *user.User
			reason                             string
			loanID, userID, deletedByID        int64
			amount, dailyInterestRate, payment float64
		}
		expectedErr error
	}{
		{
			name: "valid",
			input: struct {
				user              *user.User
				reason            string
				loanID            int64
				userID            int64
				deletedByID       int64
				amount            float64
				dailyInterestRate float64
				payment           float64
			}{
				user:              user1,
				reason:            "why not",
				loanID:            1,
				userID:            user1.ID,
				deletedByID:       user2.ID,
				amount:            100,
				dailyInterestRate: 5,
				payment:           50,
			},
		},
		{
			name: "payment = 0",
			input: struct {
				user              *user.User
				reason            string
				loanID            int64
				userID            int64
				deletedByID       int64
				amount            float64
				dailyInterestRate float64
				payment           float64
			}{
				user:              user1,
				reason:            "why not",
				loanID:            1,
				userID:            user1.ID,
				deletedByID:       user2.ID,
				amount:            100,
				dailyInterestRate: 5,
				payment:           0,
			},
			expectedErr: validator.ErrFailedValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetDB() // clean the database and start on clean slate
			userSvc := &user.Service{
				Repo: &user.Repository{DB: repo.DB},
			}
			setupUserSevice(userSvc)

			svc = &Service{
				Repo:        repo,
				UserService: userSvc,
			}
			v := validator.New()
			// step 1: create loan
			gotErr := svc.GetLoan(tc.input.user, tc.input.amount, tc.input.dailyInterestRate)
			if !checkErr(t, gotErr, tc.expectedErr, "GetLoan") {
				return
			}

			// fetch the loan from the DB
			dbLoan, gotErr := repo.GetByID(tc.input.loanID, tc.input.userID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetByID") {
				return
			}
			if dbLoan.RemainingAmount != tc.input.amount {
				t.Errorf("expected remaining = %f, got %f", tc.input.amount, dbLoan.RemainingAmount)
			}
			if dbLoan.UserID != tc.input.userID {
				t.Errorf("expected user id=%d, got %d", tc.input.userID, dbLoan.UserID)
			}

			// step 2: make payment
			_, gotErr = svc.MakePayment(v, tc.input.loanID, tc.input.userID, tc.input.payment)
			if !checkErr(t, gotErr, tc.expectedErr, "MakePayment") {
				return
			}

			// fetch again to check
			dbLoan, gotErr = repo.GetByID(tc.input.loanID, tc.input.userID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetByID 2") {
				return
			}

			if dbLoan.RemainingAmount >= tc.input.amount {
				t.Errorf("expected remaining reduced, got %f", dbLoan.RemainingAmount)
			}

			// step 3: delete loan
			loanDeletion, gotErr := svc.DeleteLoan(
				v, tc.input.loanID, tc.input.userID, tc.input.deletedByID, tc.input.reason,
			)
			if !checkErr(t, gotErr, tc.expectedErr, "DeleteLoan") {
				return
			}

			if loanDeletion.LoanID != dbLoan.ID {
				t.Errorf("expected deleted loan id %d, got %d", dbLoan.ID, loanDeletion.LoanID)
			}

			// verify loan is gone
			_, gotErr = repo.GetByID(dbLoan.ID, tc.input.userID)
			if gotErr == nil {
				t.Errorf("expected error fetching deleted loan, got nil")
			}

			// if no errors to the end and we expected an error
			if tc.expectedErr != nil {
				t.Fatalf("expected error %v, got nil", tc.expectedErr)
			}
		})
	}
}

func checkErr(t *testing.T, got, expected error, msg string) bool {
	if expected != nil {
		if got != nil && got.Error() != expected.Error() {
			t.Fatalf("%s: expected error %v, got %v", msg, expected, got)
			return false
		} else if got != nil && got.Error() == expected.Error() {
			return false
		}
	} else if got != nil {
		t.Fatalf("%s: unexpected error %v", msg, got)
		return false
	}
	return true
}
