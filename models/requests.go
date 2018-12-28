package models

type LoginRequest struct {
	Username string `json:"Username" validate:"required" binding:"required"`
	Password string `json:"Password" validate:"required" binding:"required"`
}
