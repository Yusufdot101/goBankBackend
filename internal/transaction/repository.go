package transaction

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(transaction *Transaction) error {
	query := `
		INSERT INTO transactions (user_id, action, amount, performed_by)
		VALUES ($1, $2, $3, $4)	
		RETURNING id, created_at
	`
	args := []any{
		transaction.UserID,
		transaction.Action,
		transaction.Amount,
		transaction.PerformedBy,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(
		&transaction.ID,
		&transaction.CreatedAt,
	)
}
