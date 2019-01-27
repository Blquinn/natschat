package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model

	PublicID  string `gorm:"unique;not null;"`
	Username  string `gorm:"not null;unique"`
	Password  string `gorm:"not null;"`
	Email     string `gorm:"not null;"`
	FirstName string `gorm:"not null;"`
	LastName  string `gorm:"not null;"`

	ChatMessages []*ChatMessage `gorm:"foreignkey:UserID"`
}

func (u User) ToDTO() UserDTO {
	return UserDTO{
		ID:        u.PublicID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

func (u User) ToPublicDTO() PublicUserDTO {
	return PublicUserDTO{Username: u.Username}
}

type JWTUser struct {
	ID       uint
	Username string
	Email    string
}

type UserDTO struct {
	ID        string // PublicID
	Username  string
	Email     string
	FirstName string
	LastName  string
}

type PublicUserDTO struct {
	Username string
}
