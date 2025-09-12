package tests

import (
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"

	_ "github.com/lib/pq"
)

func TestUser(t *testing.T) {
	userRepo = &user.Repository{
		DB: testDB,
	}
	tokenRepo = &token.Repository{DB: userRepo.DB}

	tokenSvc = &token.Service{
		Repo: tokenRepo,
	}
	userSvc = &user.Service{
		Repo:         userRepo,
		TokenService: tokenSvc,
	}

	user1 = &user.User{
		ID:    1,
		Name:  "yusuf",
		Email: "y@gmail.com",
	}
	user1.Password.Set("12345678", 12)

	user2 = &user.User{
		ID:             2,
		Name:           "mohamed",
		Email:          "m@gmail.com",
		AccountBalance: 100, // needed to make the tranfer money in the test
	}
	user2.Password.Set("12345678", 12)

	setupUserSevice := func(us *user.Service, user *user.User) {
		// seed the users table, this will be used in transferring of money
		us.Repo.Insert(user)
	}

	tests := []struct {
		name  string
		setup func()
		input struct {
			v              *validator.Validator
			user, fromUser *user.User
			userPassword   string
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
				v            *validator.Validator
				user         *user.User
				fromUser     *user.User
				userPassword string
				amount       float64
			}{
				user:         user1,
				userPassword: "12345678",
				fromUser:     user2,
				amount:       100,
			},
		},
		{
			name:  "duplicate email",
			setup: func() {}, // we dont reset the database so the user already exists in the db
			input: struct {
				v            *validator.Validator
				user         *user.User
				fromUser     *user.User
				userPassword string
				amount       float64
			}{
				user:         user1,
				fromUser:     user2,
				userPassword: "12345678",
				amount:       100,
			},
			expectedErr: user.ErrDuplicateEmail,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			v := validator.New()
			// step 1: register the user
			_, tkn, gotErr := userSvc.Register(
				v, tc.input.user.Name, tc.input.user.Email, tc.input.userPassword,
			)
			if !checkErr(t, gotErr, tc.expectedErr, "Register") {
				return
			}

			// fetch and check the user
			gotUser, gotErr := userSvc.GetUser(tc.input.user.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser") {
				return
			}
			if !checkUser(t, gotUser, tc.input.user, "GetUser") {
				return
			}

			// fetch the user by email and check
			gotUser, gotErr = userSvc.GetUserByEmail(tc.input.user.Email)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser") {
				return
			}
			if !checkUser(t, gotUser, tc.input.user, "GetUser") {
				return
			}

			// step 2: activate the account
			_, gotErr = userSvc.Activate(tkn.Plaintext)
			if !checkErr(t, gotErr, tc.expectedErr, "Activate") {
				return
			}

			// fetch and check the user
			gotUser, gotErr = userSvc.GetUser(tc.input.user.ID)
			if !checkErr(t, gotErr, tc.expectedErr, "GetUser 2") {
				return
			}
			if !checkActivatedUser(t, gotUser, tc.input.user, "GetUser 2") {
				return
			}

			// step 3: transfer money into the user account
			// add new account to transfer from
			setupUserSevice(userSvc, user2)
			gotUser, gotErr = userSvc.TransferMoney(tc.input.fromUser, tc.input.user, tc.input.amount)
			if !checkErr(t, gotErr, tc.expectedErr, "TransferMoney") {
				return
			}
			if !checkFromUserAfterTransfer(t, gotUser, tc.input.fromUser, "TransferMoney") {
				return
			}

			// fetch and check the user
			gotUser, gotErr = userSvc.GetUser(tc.input.user.ID)
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

func checkUser(t *testing.T, got, expected *user.User, msg string) bool {
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

func checkActivatedUser(t *testing.T, got, expected *user.User, msg string) bool {
	passed := checkUser(t, got, expected, msg)
	if !got.Activated {
		t.Errorf("%s: expected user account to be activated", msg)
		passed = false
	}
	return passed
}

func checkFromUserAfterTransfer(
	t *testing.T, got, expected *user.User, msg string,
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
	t *testing.T, got, expected *user.User, amount float64, msg string,
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
