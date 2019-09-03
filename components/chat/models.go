package chat

import (
	"natschat/components/users"
	"time"
)

type ChatRoomDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewChatRoomDTO(name, publicID string) ChatRoomDTO {
	return ChatRoomDTO{
		ID:   publicID,
		Name: name,
	}
}

type ChatMessageDTO struct {
	ID         string              `json:"id"`
	ChatRoomID string              `json:"chatRoomId"`
	CreatedAt  time.Time           `json:"createdAt"`
	Body       string              `json:"body"`
	User       users.PublicUserDTO `json:"user"`
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
	Body        string `json:"body"`
	UserID      uint   `json:"userId"`
	ChatRoomID  uint   `json:"chatRoomId"`
	MessageType string `json:"messageType"`
}

type CreateChatRoomRequest struct {
	Name string `json:"name" validate:"required,gt=3,lte=100"`
}
