package models

type PageResponse struct {
	Results interface{}
}

type LoginRequest struct {
	Username string `json:"Username" validate:"required" binding:"required"`
	Password string `json:"Password" validate:"required" binding:"required"`
}

type CreateChatRoomRequest struct {
	Name string `json:"Name" validate:"required" binding:"required"`
}
