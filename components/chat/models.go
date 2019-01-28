package chat

import (
	"natschat/components/users"
	"time"
)

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

type ChatMessageDTO struct {
	ID         string // ID
	ChatRoomID string // ID
	CreatedAt  time.Time
	Body       string

	User users.PublicUserDTO
}

func NewChatMessageDTO(id, roomID, userID, username, body string, createdAt time.Time) ChatMessageDTO {
	return ChatMessageDTO{
		ID:         id,
		ChatRoomID: roomID,
		CreatedAt:  createdAt,
		Body:       body,
		User:       users.NewPublicUserDTO(userID, username),
	}
}

type ChatMessageRequest struct {
	Body        string
	UserID      uint
	ChatRoomID  uint
	MessageType string
}

type CreateChatRoomRequest struct {
	Name string `json:"name" validate:"required,gt=3,lte=100"`
}
