package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoRecord       = errors.New("no record")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrEditConflict   = errors.New("edit conflict")
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, account_balance, activated, version
	`

	// create a 3 sec context so that the request doesnt take too long and hold the resources
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.AccountBalance,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		// this error occurs when the email already exists on the database, because we set unique
		// constraint on the email column, is its case insensitive meaning ab@c.com = AB@C.COM
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, account_balance, activated, version
		FROM users
		WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.AccountBalance,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (r *Repository) GetForToken(tokenPlaintext, scope string) (*User, error) {
	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, 
			users.account_balance, users.activated, users.version
		FROM users
		INNER JOIN tokens 
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1 
		AND tokens.scope = $2
		AND tokens.expiry > $3
	`
	// hash the plaintext using the same algorithm we used when storing
	hashedToken := sha256.Sum256([]byte(tokenPlaintext))
	args := []any{
		hashedToken[:],
		scope,
		time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.AccountBalance,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (r *Repository) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, account_balance = $4, activated = $5, 
			version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version
	`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.AccountBalance,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: key value duplicate violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict

		default:
			return err
		}
	}

	return err
}
