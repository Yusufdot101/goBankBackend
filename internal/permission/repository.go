package permission

import (
	"context"
	"database/sql"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/lib/pq"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(code Permission) error {
	query := `
		INSERT INTO permissions (code)
		VALUES ($1)
		RETURNING code
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, code).Scan(&code)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) AllForUser(userID int64) ([]Permission, error) {
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions 
		ON users_permissions.permission_id = permissions.id
		WHERE users_permissions.user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []Permission{}
	for rows.Next() {
		var permission Permission
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *Repository) Delete(code ...string) error {
	query := `
		DELETE FROM permissions
		WHERE id IN (
			SELECT id 
			FROM permissions
			WHERE code = ANY($1)
		)
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := r.DB.ExecContext(ctx, query, code)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return user.ErrNoRecord
	}

	return nil
}

func (r *Repository) Grant(userID int64, code ...string) error {
	query := `
		INSERT INTO users_permissions
		SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)
		ON CONFLICT DO NOTHING
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.DB.ExecContext(ctx, query, userID, pq.Array(code))
	return err
}

func (r *Repository) Revoke(userID int64, code ...string) error {
	query := `
		DELETE FROM users_permissions
		WHERE user_id = $1
		AND permission_id IN (
            SELECT id FROM permissions WHERE code = ANY($2)
        )
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := r.DB.ExecContext(ctx, query, userID, code)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return user.ErrNoRecord
	}

	return nil
}
