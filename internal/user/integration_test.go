package user

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/validator"

	_ "github.com/lib/pq"
)

var (
	testDB *sql.DB
	repo   *Repository
	svc    *Service
	user1  *User
	user2  *User
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

	tokenSvc := &token.Service{
		Repo: &token.Repository{DB: repo.DB},
	}

	svc = &Service{
		Repo:         repo,
		TokenService: tokenSvc,
	}

	resetDB()

	code := m.Run()
	resetDB()
	os.Exit(code)
}

func resetDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	testDB.ExecContext(ctx, `TRUNCATE tokens, users RESTART IDENTITY CASCADE`)
}

func TestUser(t *testing.T) {
	user1 = &User{
		ID:    1,
		Name:  "yusuf",
		Email: "y@gmail.com",
	}
	user1.Password.Set("12345678", 12)

	user2 = &User{
		ID:             2,
		Name:           "mohamed",
		Email:          "m@gmail.com",
		AccountBalance: 100, // needed to make the tranfer money in the test
	}
	user2.Password.Set("12345678", 12)

	setupUserSevice := func(us *Service, user *User) {
		// seed the users table, this will be used in transferring of money
		us.Repo.Insert(user)
	}

	tests := []struct {
		name  string
		setup func()
		input struct {
			v              *validator.Validator
			user, fromUser *User
			amount         float64
		}
		expectedErr error
	}{
		{
			name: "valid",
			setup: func() {
				resetDB() // clean the database and start on clean slate
			},
			input: struct {
				v        *validator.Validator
				user     *User
				fromUser *User
				amount   float64
			}{
				user:     user1,
				fromUser: user2,
				amount:   100,
			},
		},
		{
			name:  "duplicate email",
			setup: func() {}, // we dont reset the database so the user already exists in the db
			input: struct {
				v        *validator.Validator
				user     *User
				fromUser *User
				amount   float64
			}{
				user:     user1,
				fromUser: user2,
				amount:   100,
			},
			expectedErr: ErrDuplicateEmail,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			v := validator.New()
			// step 1: register the user
			_, tkn, gotErr := svc.Register(
				v, tc.input.user.Name, tc.input.user.Email, *tc.input.user.Password.plaintext,
			)
			if !checkErr(t, gotErr, tc.expectedErr, "Register") {
				return
			}

			// fetch and check the user
			gotUser, gotErr := svc.GetUser(tc.input.user.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser") {
				return
			}
			if !checkUser(t, gotUser, tc.input.user, "GetUser") {
				return
			}

			// fetch the user by email and check
			gotUser, gotErr = svc.GetUserByEmail(tc.input.user.Email)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser") {
				return
			}
			if !checkUser(t, gotUser, tc.input.user, "GetUser") {
				return
			}

			// step 2: activate the account
			_, gotErr = svc.Activate(tkn.Plaintext)
			if !checkErr(t, gotErr, tc.expectedErr, "Activate") {
				return
			}

			// fetch and check the user
			gotUser, gotErr = svc.GetUser(tc.input.user.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser 2") {
				return
			}
			if !checkActivatedUser(t, gotUser, tc.input.user, "GetUser 2") {
				return
			}

			// step 3: transfer money into the user account
			// add new account to transfer from
			setupUserSevice(svc, user2)
			gotUser, gotErr = svc.TransferMoney(tc.input.fromUser, tc.input.user, tc.input.amount)
			if !checkErr(t, gotErr, tc.expectedErr, "TransferMoney") {
				return
			}
			if !checkFromUserAfterTransfer(t, gotUser, tc.input.fromUser, "TransferMoney") {
				return
			}

			// fetch and check the user
			gotUser, gotErr = svc.GetUser(tc.input.user.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser 3") {
				return
			}
			if !checkToUserAfterTransfer(
				t, gotUser, tc.input.user, tc.input.amount, "TransferMoney 2",
			) {
				return
			}

			// if we expected an error to occur but we didnt get any
			if tc.expectedErr != nil {
				t.Fatalf("expected error %v, got nil", tc.expectedErr)
			}
		})
	}
}

func checkUser(t *testing.T, got, expected *User, msg string) bool {
	passed := true
	if got.ID != expected.ID {
		t.Errorf("%s: expected user id=%d, got id=%d", msg, expected.ID, got.ID)
		passed = false
	}
	if got.Name != expected.Name {
		t.Errorf("%s: expected user name=%s, got name=%s", msg, expected.Name, got.Name)
		passed = false
	}
	if got.Email != expected.Email {
		t.Errorf("%s: expected user email=%s, got email=%s", msg, expected.Email, got.Email)
		passed = false
	}

	return passed
}

func checkActivatedUser(t *testing.T, got, expected *User, msg string) bool {
	passed := checkUser(t, got, expected, msg)
	if !got.Activated {
		t.Errorf("%s: expected user account to be activated", msg)
		passed = false
	}
	return passed
}

func checkFromUserAfterTransfer(
	t *testing.T, got, expected *User, msg string,
) bool {
	passed := checkUser(t, got, expected, msg)
	if got.AccountBalance != 0 {
		t.Errorf(
			"%s: expected user account balance=0, got account balance=%f", msg, got.AccountBalance,
		)
		passed = false
	}
	return passed
}

func checkToUserAfterTransfer(
	t *testing.T, got, expected *User, amount float64, msg string,
) bool {
	passed := checkUser(t, got, expected, msg)
	if got.AccountBalance != amount {
		t.Errorf(
			"%s: expected user account balance=%f, got account balance=%f", msg,
			amount, got.AccountBalance,
		)
		passed = false
	}
	return passed
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
