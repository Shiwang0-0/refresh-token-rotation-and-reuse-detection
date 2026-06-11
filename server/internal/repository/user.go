package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/models"
)

type UserRepository interface {
	IsEmailAlreadyTaken(string) (bool, error)
	RegisterUser(string, string, string) (*models.User, error)
	UserFindByEmail(string) (*models.User, error)
	SaveRefreshToken(int, string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository { // returns interface, not *struct
	return &userRepository{db: db}
}

func (r *userRepository) IsEmailAlreadyTaken(email string) (bool, error) {
	query := `SELECT id from users WHERE email = ?`

	var id int
	err := r.db.QueryRow(query, email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error selecting user: %w", err)
	}

	return true, nil
}

func (r *userRepository) RegisterUser(name, email, passwordHash string) (*models.User, error) {
	query := `INSERT into users(name, email, passwordHash) VALUES (?,?,?)`

	result, err := r.db.Exec(query, name, email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting inserted id: %w", err)
	}

	user := &models.User{
		ID:    int(id),
		Name:  name,
		Email: email,
	}

	return user, nil
}

func (r *userRepository) UserFindByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, passwordHash, refreshToken FROM users WHERE email = ?`

	var user models.User
	var refreshToken sql.NullString

	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error finding user by email: %w", err)
	}
	if refreshToken.Valid {
		user.RefreshToken = refreshToken.String
	}

	return &user, nil
}

func (r *userRepository) SaveRefreshToken(userID int, refreshToken string) error {
	query := `UPDATE users SET refreshToken = ? WHERE id = ?`

	result, err := r.db.Exec(query, refreshToken, userID)
	if err != nil {
		return fmt.Errorf("error updating refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
