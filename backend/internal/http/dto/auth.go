package dto

import "gitlab.com/docuemnt_manegment/backend/internal/domain"

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	DisplayName string `json:"display_name" validate:"max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateMeRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

func UserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:          u.ID.Hex(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
	}
}
