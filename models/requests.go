package models

type PageResponse struct {
	Results interface{}
}

type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,gt=4,lte=40"`
	Password  string `json:"password" validate:"required,gte=8"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required" binding:"required"`
	Password string `json:"password" validate:"required" binding:"required"`
}

type CreateChatRoomRequest struct {
	Name string `json:"name" validate:"required,gt=3,lte=100"`
}
