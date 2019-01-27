package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type ChatRoom struct {
	gorm.Model

	Name          string              `gorm:"unique;not null"`
	Owner         *User               `gorm:"association_foreignkey:ID"`
	OwnerID       uint                `gorm:"not null;"`
	Subscriptions []*ChatSubscription `gorm:"foreignkey:ChatRoomID"`
	Messages      []*ChatMessage      `gorm:"foreignkey:ChatRoomID"`
}

type ChatSubscription struct {
	gorm.Model

	User       *User     `gorm:"association_foreignkey:ID"`
	UserID     string    `gorm:"not null;"`
	ChatRoom   *ChatRoom `gorm:"association_foreignkey:ID"`
	ChatRoomID string    `gorm:"not null;"`
}

type ChatMessage struct {
	gorm.Model

	PublicID    string    `gorm:"not null;"`
	Body        string    `gorm:"not null;"`
	User        *User     `gorm:"association_foreignkey:ID"`
	UserID      string    `gorm:"not null;"`
	ChatRoom    *ChatRoom `gorm:"association_foreignkey:ID"`
	ChatRoomID  string    `gorm:"not null;"`
	MessageType string    `gorm:"not null;"`
}

type ChatMessageDTO struct {
	ID         string // PublicID
	ChatRoomID string // PublicID
	CreatedAt  time.Time
	Body       string

	User PublicUserDTO
}

func NewChatMessageDTO(id, roomID, username, body string, createdAt time.Time) ChatMessageDTO {
	return ChatMessageDTO{
		ID:         id,
		ChatRoomID: roomID,
		CreatedAt:  createdAt,
		Body:       body,
		User: PublicUserDTO{
			Username: username,
		},
	}
}
