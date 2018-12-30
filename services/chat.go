package services

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"natschat/models"
	"natschat/utils"
	"time"
)

type IChatService interface {
	CreateChatRoom(name string) (models.ChatRoom, *utils.APIError)
	ListChatRooms() ([]models.ChatRoom, *utils.APIError)
	GetChatHistory(roomID string) ([]models.ChatMessageDTO, *utils.APIError)
}

type ChatService struct {
	db *sqlx.DB
}

var _ IChatService = &ChatService{}

func NewChatService(db *sqlx.DB) *ChatService {
	return &ChatService{
		db: db,
	}
}

func (cs *ChatService) CreateChatRoom(name string) (models.ChatRoom, *utils.APIError) {
	id, _ := uuid.NewV4()
	t := time.Now()
	_, err := cs.db.Exec(`
		insert into chat_rooms (id, "name", inserted_at, updated_at)	
		values ($1, $2, $3, $3)
	`, id.String(), name, t)

	room := models.ChatRoom{}
	if err != nil {
		if isDuplicateError(err) {
			return room, utils.NewPubicError(err, "A chat room with that name already exists.")
		}
		logrus.Errorf("Error occurred while saving chat room: %v", err)
		return room, utils.NewPrivateError(err)
	}

	return models.ChatRoom{
		ID: id.String(),
		Name: name,
		InsertedAt: t,
		UpdatedAt: t,
	}, nil

}

func (cs *ChatService) ListChatRooms() ([]models.ChatRoom, *utils.APIError) {
	var rooms []models.ChatRoom
	err := cs.db.Select(&rooms, `
		select id, "name", inserted_at, updated_at
		from chat_rooms
	`)
	if err != nil {
		logrus.Errorf("Error occurred while listing chat rooms: %v", err)
		return rooms, &utils.APIError{err, err.Error(), false}
	}

	return rooms, nil
}

func (cs *ChatService) GetChatHistory(roomID string) ([]models.ChatMessageDTO, *utils.APIError) {
	var msgs []models.ChatMessagePlusUser
	err := cs.db.Select(&msgs, `
		select
			   cm.*,
		       uu.id "user.id",
		       uu.username "user.username",
		       uu.email "user.email",
		       uu.first_name "user.first_name",
		       uu.last_name "user.last_name",
		       uu.inserted_at "user.inserted_at",
		       uu.updated_at "user.updated_at"
		from chat_messages cm
		inner join users_user uu on cm.user_id = uu.id
		where chat_room_id = $1
	`, roomID)
	if err != nil {
		return []models.ChatMessageDTO{}, utils.NewPrivateError(err)
	}

	if msgs == nil {
		return []models.ChatMessageDTO{}, nil
	}

	dtos := make([]models.ChatMessageDTO, len(msgs))
	for i, m := range msgs {
		dtos[i] = models.ChatMessageDTO{
			ID:         m.ID,
			InsertedAt: m.InsertedAt,
			Body:       m.Body,
			User: models.PublicUserDTO{
				Username: m.User.Username,
			},
		}
	}

	return dtos, nil
}

// TODO: this
//func (cs *ChatService) SaveMessage(roomID string) ([]models.ChatMessage, *utils.APIError) {
//	var msgs []models.ChatMessage
//	err := cs.db.Select(&msgs, `
//		select id, inserted_at, updated_at, user_id, chat_room_id, body, message_type
//		from chat_messages
//		where chat_room_id = $1
//	`, roomID)
//	if err != nil {
//		return msgs, utils.NewPrivateError(err)
//	}
//
//	return msgs, nil
//}

func isDuplicateError(err error) bool {
	if e, ok := err.(*pq.Error); ok && e.Code == "23505" {
		return true
	}
	return false
}
