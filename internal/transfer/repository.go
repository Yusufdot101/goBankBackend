package transfer

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(transfer *Transfer) error {
	query := `
		INSERT INTO transfers (from_user_id, to_user_id, amount)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(
		ctx, query,
		transfer.FromUserID,
		transfer.ToUserID,
		transfer.Amount,
	).Scan(&transfer.ID)
}
