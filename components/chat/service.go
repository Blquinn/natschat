package chat

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"natschat/models"
	"natschat/utils/apierrs"
	"natschat/utils/auth"
	"natschat/utils/db"
	"net/http"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

func (cs *Service) CreateChatRoom(name string, userID uint) (ChatRoomDTO, *apierrs.APIError) {
	var rd ChatRoomDTO
	r := models.NewChatRoom(name, userID)
	if err := cs.db.Create(&r).Error; err != nil {
		if db.IsDuplicateError(err) {
			return rd, apierrs.NewPublicError(err, "A chat room with that name already exists.", http.StatusBadRequest)
		}
		return rd, apierrs.NewPrivateError(err)
	}
	return NewChatRoomDTO(r.Name, r.PublicID), nil
}

func (cs *Service) ListChatRooms() ([]ChatRoomDTO, *apierrs.APIError) {
	var rs []models.ChatRoom
	if err := cs.db.Find(&rs).Error; err != nil {
		return []ChatRoomDTO{}, apierrs.NewPrivateError(err)
	}

	rsd := make([]ChatRoomDTO, len(rs))
	for i, r := range rs {
		rsd[i] = NewChatRoomDTO(r.Name, r.PublicID)
	}
	return rsd, nil
}

func (cs *Service) GetChatHistory(roomID string) ([]ChatMessageDTO, *apierrs.APIError) {
	var msgs []models.ChatMessage
	if err := cs.db.
		Preload("User").
		Preload("ChatRoom").
		Joins("inner join chat_rooms on chat_rooms.id = chat_room_id").
		Where("chat_rooms.public_id = ?", roomID).
		Find(&msgs).Error; err != nil {
		return []ChatMessageDTO{}, apierrs.NewPrivateError(err)
	}

	msgDTOs := make([]ChatMessageDTO, len(msgs))
	for i, m := range msgs {
		msgDTOs[i] = NewChatMessageDTO(m.PublicID, m.ChatRoom.PublicID, m.User.PublicID, m.User.Username, m.Body, m.CreatedAt)
	}
	return msgDTOs, nil
}

func (cs *Service) SaveChatMessage(body, chatRoomID string, user *auth.JWTUser) (ChatMessageDTO, *apierrs.APIError) {
	var r models.ChatRoom
	if err := cs.db.Where("public_id = ?", chatRoomID).First(&r).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ChatMessageDTO{}, apierrs.NewPublicError(err,
				fmt.Sprintf("Chat room with id `%s` not found", chatRoomID),
				http.StatusBadRequest)
		}
		log.Errorf("Error occurred while getting chat room: %v", err)
		return ChatMessageDTO{}, apierrs.NewPrivateError(err)
	}

	m := models.NewChatMessage(body, user.ID, r.ID)
	m.PublicID = uuid.NewV4().String()
	if err := cs.db.Create(&m).Error; err != nil {
		return ChatMessageDTO{}, apierrs.NewPrivateError(err)
	}
	return NewChatMessageDTO(m.PublicID, r.PublicID, user.PublicID, user.Username, m.Body, m.CreatedAt), nil
}
