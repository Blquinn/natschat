package users

import "natschat/models"

type UserDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func NewUserDTO(id, username, email, firstName, lastName string) UserDTO {
	return UserDTO{
		ID:        id,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
}

func UserToDTO(user models.User) UserDTO {
	return UserDTO{
		ID:        user.PublicID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

type PublicUserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewPublicUserDTO(id, username string) PublicUserDTO {
	return PublicUserDTO{
		ID:       id,
		Username: username,
	}
}

type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,gte=3,lte=40"`
	Password  string `json:"password" validate:"required,gte=8"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required" binding:"required"`
	Password string `json:"password" validate:"required" binding:"required"`
}

type AuthResponseDTO struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}
