package models

import "time"

type ChatRoom struct {
	ID string `db:"id"`
	InsertedAt time.Time `db:"inserted_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Name string `db:"name"`
}

type ChatSubscription struct {
	ID string `db:"id"`
	InsertedAt time.Time `db:"inserted_at"`
	UpdatedAt time.Time `db:"updated_at"`

	UserID string `db:"user_id"`
	ChatRoomID string `db:"chat_room_id"`
}

type ChatMessage struct {
	ID string `db:"id"`
	InsertedAt time.Time `db:"inserted_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Body string `db:"body"`
	UserID string `db:"user_id"`
	ChatRoomID string `db:"chat_room_id"`
	MessageType string `db:"message_type"`
}

type ChatMessagePlusUser struct {
	ID string `db:"id"`
	InsertedAt time.Time `db:"inserted_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Body string `db:"body"`
	UserID string `db:"user_id"`
	ChatRoomID string `db:"chat_room_id"`
	MessageType string `db:"message_type"`

	User User `db:"user"`
}

type ChatMessageDTO struct {
	ID string
	InsertedAt time.Time
	Body string

	User PublicUserDTO
}
