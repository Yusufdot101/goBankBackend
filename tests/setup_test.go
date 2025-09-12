package tests

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/loan"
	"github.com/Yusufdot101/goBankBackend/internal/loanrequests"
	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/user"
)

var (
	testDB *sql.DB

	userRepo  *user.Repository
	tokenRepo *token.Repository
	// permissionRepo  *permission.Repository
	loanrequestRepo *loanrequests.Repository
	loanRepo        *loan.Repository
	// transferRepo    *transfer.Repository
	// transactionRepo *transaction.Repository

	userSvc  *user.Service
	tokenSvc *token.Service
	// permissionSvc  *permission.Service
	loanrequestSvc *loanrequests.Service
	loanSvc        *loan.Service
	// transferSvc    *transfer.Service
	// transactionSvc *transaction.Service

	user1 *user.User
	user2 *user.User
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

	resetDB()

	code := m.Run()
	resetDB()
	os.Exit(code)
}

func resetDB() {
	query := `
		TRUNCATE loans, deleted_loans, loan_requests, permissions, users_permissions, tokens, 
			transactions, transfers, users RESTART IDENTITY CASCADE
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := testDB.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
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
