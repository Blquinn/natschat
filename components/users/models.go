package users

type UserDTO struct {
	ID        string // ID
	Username  string
	Email     string
	FirstName string
	LastName  string
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

type PublicUserDTO struct {
	ID       string // ID
	Username string
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
