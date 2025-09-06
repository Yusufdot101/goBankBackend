package loan

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(loan *Loan) error {
	query := `
		INSERT INTO loans 
			(user_id, amount, action, daily_interest_rate, remaining_amount, last_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	args := []any{
		loan.UserID,
		loan.Amount,
		loan.Action,
		loan.DailyInterestRate,
		loan.RemainingAmount,
		loan.LastUpdatedAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(
		&loan.ID,
		&loan.CreatedAt,
	)
}

func (r *Repository) GetByID(loanID, userID int64) (*Loan, error) {
	query := `
		SELECT id, created_at, user_id, amount, action, daily_interest_rate, remaining_amount, 
			last_updated_at, version
		FROM loans
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var loan Loan
	err := r.DB.QueryRowContext(ctx, query, loanID, userID).Scan(
		&loan.ID,
		&loan.CreatedAt,
		&loan.UserID,
		&loan.Amount,
		&loan.Action,
		&loan.DailyInterestRate,
		&loan.RemainingAmount,
		&loan.LastUpdatedAt,
		&loan.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, user.ErrNoRecord
		default:
			return nil, err
		}
	}

	return &loan, nil
}

// GetForUserByUserID gets all loans and payments for a give user by their ID
func (r *Repository) GetForUserByUserID(userID int64) ([]*Loan, error) {
	query := `
		SELECT id, created_at, user_id, amount, action, daily_interest_rate, remaining_amount, 
			version
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []*Loan
	for rows.Next() {
		var loan Loan
		err = rows.Scan(
			&loan.ID,
			&loan.CreatedAt,
			&loan.UserID,
			&loan.Amount,
			&loan.Action,
			&loan.DailyInterestRate,
			&loan.Amount,
			&loan.Version,
		)
		if err != nil {
			return nil, err
		}

		loans = append(loans, &loan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}

func (r *Repository) MakePaymentTx(loanID, userID int64, payment, totalOwed float64) (*Loan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// rollback if anything goes wrong
	defer tx.Rollback()

	// fetch loan with FOR UPDATE to lock the row
	loan := &Loan{}
	query := `
		SELECT id, user_id, remaining_amount, daily_interest_rate, last_updated_at
		FROM loans
		WHERE id = $1 AND user_id = $2
		FOR UPDATE 
	`

	err = tx.QueryRowContext(ctx, query, loanID, userID).Scan(
		&loan.ID,
		&loan.UserID,
		&loan.RemainingAmount,
		&loan.DailyInterestRate,
		&loan.LastUpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, user.ErrNoRecord
		default:
			return nil, err
		}
	}

	loan.RemainingAmount = math.Max(0, totalOwed-payment)
	loan.LastUpdatedAt = time.Now().UTC()

	// update the row in the database
	updateQuery := `
		UPDATE loans
		SET remaining_amount = $1, last_updated_at = $2, version = version + 1
		WHERE id = $3 AND user_id = $4
	`
	args := []any{
		loan.RemainingAmount,
		loan.LastUpdatedAt,
		loan.ID,
		loan.UserID,
	}

	_, err = tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return loan, nil
}

func (r *Repository) DeleteLoan(loanID, userID int64) error {
	query := `
		DELETE FROM loans
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := r.DB.ExecContext(ctx, query, loanID, userID)
	if err != nil {
		return err
	}

	rowsEffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsEffected == 0 {
		return user.ErrNoRecord
	}

	return nil
}

func (r *Repository) InsertDeletion(loanDeletion *LoanDeletion) error {
	query := `
		INSERT INTO deleted_loans 
		(
			loan_created_at, loan_last_updated_at, loan_id, debtor_id, deleted_by_id, amount, 
			daily_interest_rate, remaining_amount, reason
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	args := []any{
		loanDeletion.LoanCreatedAt,
		loanDeletion.LoanLastUpdatedAt,
		loanDeletion.LoanID,
		loanDeletion.DebtorID,
		loanDeletion.DeletedByID,
		loanDeletion.Amount,
		loanDeletion.DailyInterestRate,
		loanDeletion.RemainingAmount,
		loanDeletion.Reason,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(
		&loanDeletion.ID,
		&loanDeletion.CreatedAt,
	)
}
