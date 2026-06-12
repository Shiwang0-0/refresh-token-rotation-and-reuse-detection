package services

import (
	"errors"
	"fmt"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/dto"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/repository"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/utils"
)

type UserService interface {
	RegisterUser(dto.RegisterRequest) (*dto.UserProfileResponse, error)
	LoginUser(dto.LoginRequest) (*dto.UserProfileResponse, string, string, error)
	GetUserProfile(int) (*dto.UserProfileResponse, error)
	RefreshTokens(string) (string, string, error)
	LogoutUser(string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) RegisterUser(data dto.RegisterRequest) (*dto.UserProfileResponse, error) {

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

	return &dto.UserProfileResponse{
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *userService) LoginUser(data dto.LoginRequest) (*dto.UserProfileResponse, string, string, error) {
	fmt.Printf("data: %v", data)
	user, err := s.userRepo.UserFindByEmail(data.Email)

	if err != nil {
		return nil, "", "", utils.ErrInvalidCredentials
	}

	if err := utils.CheckPasswordHash(user.PasswordHash, data.Password); err != nil {
		return nil, "", "", utils.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	fmt.Println("ACCESS TOKEN: ", accessToken)
	fmt.Println("REFRESH TOKEN: ", refreshToken)

	if err := s.userRepo.SaveRefreshToken(user.ID, refreshToken); err != nil {
		return nil, "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &dto.UserProfileResponse{
		Name:  user.Name,
		Email: user.Email,
	}, accessToken, refreshToken, nil
}

func (s *userService) GetUserProfile(userID int) (*dto.UserProfileResponse, error) {

	user, err := s.userRepo.GetUserProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching user profile")
	}

	return &dto.UserProfileResponse{
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *userService) RefreshTokens(refreshToken string) (string, string, error) {
	userID, err := utils.ValidateRefreshToken(refreshToken)

	if err != nil {
		// additional: set the refresh token to NULL, if the
		fmt.Println("INVALID TOKEN")
		if errors.Is(err, utils.ErrTokenExpired) {
			fmt.Println("TOKEN EXPIRED")
			// only clean up on expiry, not on invalid token
			_ = s.userRepo.ClearRefreshToken(refreshToken)
		}
		return "", "", utils.ErrInvalidCredentials
	}

	newAccessToken, err := utils.GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	fmt.Println("NEW ACCESS TOKEN: ", newAccessToken)
	fmt.Println("NEW REFRESH TOKEN: ", newRefreshToken)

	fmt.Printf("USER: %d\n", userID)
	if err := s.userRepo.SaveRefreshToken(userID, newRefreshToken); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *userService) LogoutUser(refreshToken string) error {
	// idemopotent logout
	// even if token is expired, invalid just delete without validating
	if err := s.userRepo.ClearRefreshToken(refreshToken); err != nil {
		return fmt.Errorf("failed to clear refresh token: %w", err)
	}
	return nil
}
