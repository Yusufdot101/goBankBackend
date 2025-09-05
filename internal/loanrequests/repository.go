package loanrequests

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(loanRequest *LoanRequest) error {
	query := `
		INSERT INTO loan_requests
			(user_id, amount, daily_interest_rate, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	args := []any{
		loanRequest.UserID,
		loanRequest.Amount,
		loanRequest.DailyInterestRate,
		loanRequest.Status,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(
		&loanRequest.ID,
		&loanRequest.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Get(loanRequestID, userID int64) (*LoanRequest, error) {
	query := `
		SELECT id, created_at, user_id, amount, daily_interest_rate, status
		WHERE id = $1
		AND user_id = $2
	`
	loanRequest := &LoanRequest{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, loanRequestID, userID).Scan(
		&loanRequest.ID,
		&loanRequest.CreatedAt,
		&loanRequest.UserID,
		&loanRequest.Amount,
		&loanRequest.DailyInterestRate,
		&loanRequest.Status,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, user.ErrNoRecord
		default:
			return nil, err
		}
	}

	return loanRequest, nil
}

func (r *Repository) UpdateTx(loanRequestID, userID int64, newStatus string) (*LoanRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// in case of any issues
	defer tx.Rollback()

	// fetch loan request, use FOR UPDATE to lock the row from others trying to update at the same
	// time
	query := `
		SELECT id, created_at, user_id, amount, daily_interest_rate, status 
		FROM  loan_requests
		WHERE id = $1 
		AND user_id = $2
		FOR UPDATE
	`
	loanRequest := &LoanRequest{}
	err = tx.QueryRowContext(ctx, query, loanRequestID, userID).Scan(
		&loanRequest.ID,
		&loanRequest.CreatedAt,
		&loanRequest.UserID,
		&loanRequest.Amount,
		&loanRequest.DailyInterestRate,
		&loanRequest.Status,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, user.ErrNoRecord
		default:
			return nil, err
		}
	}

	updateQuery := `
		UPDATE loan_requests
		SET status = $1
		WHERE id = $2
		AND user_id = $3
	`

	_, err = tx.ExecContext(ctx, updateQuery, newStatus, loanRequest.ID, loanRequest.UserID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return loanRequest, nil
}
