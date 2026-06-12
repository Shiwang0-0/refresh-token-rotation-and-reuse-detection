package dto

type RegisterRequest struct {
	Name     string `json:"name"      validate:"required,min=2,max=50"`
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=6,max=50"`
}

type RegisterResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=6,max=50"`
}

type UserProfileResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
