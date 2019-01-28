package services

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"natschat/models"
	"natschat/utils"
	"net/http"
)

type IChatService interface {
	CreateChatRoom(name string, userID uint) (models.ChatRoomDTO, *utils.APIError)
	ListChatRooms() ([]models.ChatRoom, *utils.APIError)
	GetChatHistory(roomID string) ([]models.ChatMessageDTO, *utils.APIError)
	SaveChatMessage(message models.ChatMessage) (models.ChatMessage, error)
}

type ChatService struct {
	db *gorm.DB
}

var _ IChatService = &ChatService{}

func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{
		db: db,
	}
}

func (cs *ChatService) CreateChatRoom(name string, userID uint) (models.ChatRoomDTO, *utils.APIError) {
	var rd models.ChatRoomDTO
	r := models.NewChatRoom(name, userID)
	if err := cs.db.Create(&r).Error; err != nil {
		if isDuplicateError(err) {
			return rd, utils.NewPublicError(err, "A chat room with that name already exists.", http.StatusBadRequest)
		}
		return rd, utils.NewPrivateError(err)
	}
	return models.NewChatRoomDTO(r.Name, r.PublicID), nil
}

func (cs *ChatService) ListChatRooms() ([]models.ChatRoom, *utils.APIError) {
	var rs []models.ChatRoom
	if err := cs.db.Find(&rs).Error; err != nil {
		return rs, utils.NewPrivateError(err)
	}
	return rs, nil
}

func (cs *ChatService) GetChatHistory(roomID string) ([]models.ChatMessageDTO, *utils.APIError) {
	var msgs []models.ChatMessage
	if err := cs.db.
		Preload("User").
		Preload("ChatRoom").
		Where("chat_room_id = ?", roomID).Find(&msgs).Error; err != nil {
		return []models.ChatMessageDTO{}, utils.NewPrivateError(err)
	}

	msgDTOs := make([]models.ChatMessageDTO, len(msgs))
	for i, m := range msgs {
		msgDTOs[i] = models.NewChatMessageDTO(m.PublicID, m.ChatRoom.PublicID, m.User.PublicID, m.User.Username, m.Body, m.CreatedAt)
	}
	return msgDTOs, nil
}

func (cs *ChatService) SaveChatMessage(m models.ChatMessage) (models.ChatMessage, error) {
	m.PublicID = uuid.NewV4().String()
	if err := cs.db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func isDuplicateError(err error) bool {
	if e, ok := err.(*pq.Error); ok && e.Code == "23505" {
		return true
	}
	return false
}
