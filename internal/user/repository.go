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
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, account_balance)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, activated, version
	`

	// create a 3 sec context so that the request doesnt take too long and hold the resources
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(
		ctx, query, user.Name, user.Email, user.Password.Hash, user.AccountBalance,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		// &user.AccountBalance,
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

func (r *Repository) Get(userID int64) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, account_balance, activated, version
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := r.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
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
		&user.Password.Hash,
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
		&user.Password.Hash,
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

func (r *Repository) UpdateTx(
	userID int64, name, email string, passwordHash []byte, accountBalance float64, activated bool,
) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, created_at, name, email, password_hash, account_balance, activated, version
		FROM users
		WHERE id = $1
		FOR UPDATE
	`

	user := &User{}
	err = tx.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.AccountBalance,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		return nil, err
	}

	updateQuery := `
		UPDATE users
		set name = $1, email = $2, password_hash = $3, account_balance = $4, activated = $5, 
			version = version + 1
		WHERE id = $6
		RETURNING name, email, password_hash, account_balance, activated
	`

	args := []any{
		name,
		email,
		passwordHash,
		accountBalance,
		activated,
		userID,
	}

	err = tx.QueryRowContext(ctx, updateQuery, args...).Scan(
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.AccountBalance,
		&user.Activated,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}
