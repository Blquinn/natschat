package models

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type User struct {
	gorm.Model

	PublicID  string `gorm:"unique;not null;"`
	Username  string `gorm:"not null;unique"`
	Password  string `gorm:"not null;"`
	Email     string `gorm:"not null;"`
	FirstName string `gorm:"not null;"`
	LastName  string `gorm:"not null;"`

	ChatMessages []*ChatMessage `gorm:"foreignkey:UserID"`
	ChatRooms    []*ChatRoom    `gorm:"foreignkey:OwnerID"`
}

func NewUser(username, password, email, firstName, lastName string) User {
	return User{
		PublicID:  uuid.NewV4().String(),
		Username:  username,
		Password:  password,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
}
