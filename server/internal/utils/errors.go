package utils

import "errors"

var (
	ErrEmailTaken         = errors.New("account with this email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
