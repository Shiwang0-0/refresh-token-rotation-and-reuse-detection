package models

import "time"

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
	RefreshToken string // empty string means not logged in
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
