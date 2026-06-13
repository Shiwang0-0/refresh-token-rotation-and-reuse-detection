package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/models"
)

type UserRepository interface {
	IsEmailAlreadyTaken(string) (bool, error)
	RegisterUser(string, string, string) (*models.User, error)
	UserFindByEmail(string) (*models.User, error)
	GetUserProfile(int) (*models.User, error)

	SaveRefreshToken(userID int, refreshToken string, expiresAt time.Time) error
	GetUserByRefreshToken(refreshToken string) (*models.User, error)
	DeleteRefreshToken(refreshToken string) error // logout / rotation
	DeleteAllUserSessions(userID int) error       // reuse detection
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
	query := `SELECT id, name, email, passwordHash FROM users WHERE email = ?`

	var user models.User

	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error finding user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) SaveRefreshToken(userID int, refreshToken string, expiresAt time.Time) error {
	query := `INSERT INTO sessions(userId, refreshToken, expiresAt) VALUES (?,?,?)`
	_, err := r.db.Exec(query, userID, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("SaveRefreshToken: %w", err)
	}
	return nil
}

func (r *userRepository) GetUserProfile(userID int) (*models.User, error) {
	query := `SELECT name, email from users WHERE id = ?`

	var user models.User
	err := r.db.QueryRow(query, userID).Scan(&user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error finding user profile: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetUserByRefreshToken(refreshToken string) (*models.User, error) {
	query := `
        SELECT u.id, u.name, u.email FROM users u
        INNER JOIN 
		sessions s ON s.userId = u.id
        WHERE s.refreshToken = ? AND s.expiresAt > NOW()`

	var user models.User
	err := r.db.QueryRow(query, refreshToken).Scan(&user.ID, &user.Name, &user.Email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetUserByRefreshToken: %w", err)
	}
	return &user, nil
}

func (r *userRepository) DeleteRefreshToken(refreshToken string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE refreshToken = ?`, refreshToken)
	if err != nil {
		return fmt.Errorf("DeleteRefreshToken: %w", err)
	}
	return nil
}

// DeleteAllUserSessions deletes every session for a user (reuse detection)
func (r *userRepository) DeleteAllUserSessions(userID int) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE userId = ?`, userID)
	if err != nil {
		return fmt.Errorf("DeleteAllUserSessions: %w", err)
	}
	return nil
}
