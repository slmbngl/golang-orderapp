package models

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	Role         string    `json:"role" db:"role"`
}
type GetMeResponseReq struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	Role      string    `json:"role"`
}
type UserLoginResponseReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserRegisterResponseReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
type UserResponseReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin" example:"admin"`
}

// RefreshTokenRequest represents the request body for refresh token endpoint
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// LogoutRequest represents the request body for logout endpoint
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResponse represents the response for refresh token endpoint
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in" example:"900"`
}
