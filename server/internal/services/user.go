package services

import (
	"fmt"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/dto"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/repository"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/utils"
)

type UserService interface {
	RegisterUser(dto.RegisterRequest) (*dto.RegisterResponse, error)
	LoginUser(dto.LoginRequest) (string, string, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) RegisterUser(data dto.RegisterRequest) (*dto.RegisterResponse, error) {

	isEmailAlreadyTaken, err := s.userRepo.IsEmailAlreadyTaken(data.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email presence: %w", err)
	}
	if isEmailAlreadyTaken {
		return nil, fmt.Errorf("Account with this email already exist")
	}

	// hash password
	passwordHash, err := utils.HashPassword(data.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// register the user
	user, err := s.userRepo.RegisterUser(data.Name, data.Email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return &dto.RegisterResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *userService) LoginUser(data dto.LoginRequest) (string, string, error) {
	fmt.Printf("data: %v", data)
	user, err := s.userRepo.UserFindByEmail(data.Email)

	if err != nil {
		fmt.Print("Error finding email ??")
		return "", "", utils.ErrInvalidCredentials
	}

	if err := utils.CheckPasswordHash(user.PasswordHash, data.Password); err != nil {
		fmt.Print("Error checking password hash ??")
		return "", "", utils.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	fmt.Println("ACCESS TOKEN: ", accessToken)
	fmt.Println("REFRESH TOKEN: ", refreshToken)

	if err := s.userRepo.SaveRefreshToken(user.ID, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
