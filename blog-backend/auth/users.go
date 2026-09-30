package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrInvalidCredentials is returned when the username is unknown or the
// password does not match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Users is the blog_users table.
type Users struct {
	db *sql.DB
}

// NewUsers returns a Users backed by db.
func NewUsers(db *sql.DB) *Users {
	return &Users{db: db}
}

// Authenticate checks username and password against the stored hash.
func (u *Users) Authenticate(ctx context.Context, username, password string) error {
	var hash string
	err := u.db.QueryRowContext(ctx,
		"SELECT password_hash FROM blog_users WHERE username = ?", username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("look up user: %w", err)
	}
	if !CheckPasswordHash(password, hash) {
		return ErrInvalidCredentials
	}
	return nil
}

// Count returns the number of registered users.
func (u *Users) Count(ctx context.Context) (int, error) {
	var count int
	err := u.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM blog_users").Scan(&count)
	return count, err
}

// Create registers a user with the given plain-text password.
func (u *Users) Create(ctx context.Context, username, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = u.db.ExecContext(ctx,
		"INSERT INTO blog_users (username, password_hash) VALUES (?, ?)", username, hash)
	return err
}
