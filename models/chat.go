package models

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"time"
)

type ChatRoom struct {
	gorm.Model

	Name          string              `gorm:"unique;not null"`
	PublicID      string              `gorm:"unique;not null;"`
	Owner         *User               `gorm:"association_foreignkey:ID"`
	OwnerID       uint                `gorm:"not null;"`
	Subscriptions []*ChatSubscription `gorm:"foreignkey:ChatRoomID"`
	Messages      []*ChatMessage      `gorm:"foreignkey:ChatRoomID"`
}

func NewChatRoom(name string, ownerID uint) ChatRoom {
	return ChatRoom{
		Name:     name,
		PublicID: uuid.NewV4().String(),
		OwnerID:  ownerID,
	}
}

type ChatRoomDTO struct {
	Name string
	ID   string
}

func NewChatRoomDTO(name, publicID string) ChatRoomDTO {
	return ChatRoomDTO{
		Name: name,
		ID:   publicID,
	}
}

type ChatSubscription struct {
	gorm.Model

	User       *User     `gorm:"association_foreignkey:ID"`
	UserID     uint      `gorm:"not null;"`
	ChatRoom   *ChatRoom `gorm:"association_foreignkey:ID"`
	ChatRoomID uint      `gorm:"not null;"`
}

func NewChatSubscription(userID, chatRoomID uint) ChatSubscription {
	return ChatSubscription{
		UserID:     userID,
		ChatRoomID: chatRoomID,
	}
}

type ChatMessage struct {
	gorm.Model

	PublicID    string    `gorm:"not null;"`
	Body        string    `gorm:"not null;"`
	User        *User     `gorm:"association_foreignkey:ID"`
	UserID      uint      `gorm:"not null;"`
	ChatRoom    *ChatRoom `gorm:"association_foreignkey:ID"`
	ChatRoomID  uint      `gorm:"not null;"`
	MessageType string    `gorm:"not null;"`
}

func NewChatMessage(body, messageType string, userID, chatRoomID uint) ChatMessage {
	return ChatMessage{
		PublicID:    uuid.NewV4().String(),
		Body:        body,
		UserID:      userID,
		ChatRoomID:  chatRoomID,
		MessageType: messageType,
	}
}

type ChatMessageDTO struct {
	ID         string // ID
	ChatRoomID string // ID
	CreatedAt  time.Time
	Body       string

	User PublicUserDTO
}

func NewChatMessageDTO(id, roomID, userID, username, body string, createdAt time.Time) ChatMessageDTO {
	return ChatMessageDTO{
		ID:         id,
		ChatRoomID: roomID,
		CreatedAt:  createdAt,
		Body:       body,
		User:       NewPublicUserDTO(userID, username),
	}
}
