package services

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"natschat/models"
	"natschat/utils"
	"net/http"
)

type IUserService interface {
	GetUserByUsername(username string) (models.User, error)
	CreateUser(userRequest models.CreateUserRequest) (models.User, *utils.APIError)
}

var _ IUserService = &UserService{}

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (us *UserService) GetUserByUsername(username string) (models.User, error) {
	u := models.User{}
	if err := us.db.First(&u, "username = ?", username).Error; err != nil {
		return u, err
	}

	return u, nil
}

func (us *UserService) CreateUser(ur models.CreateUserRequest) (models.User, *utils.APIError) {
	u := models.User{
		PublicID:  uuid.NewV4().String(),
		Username:  ur.Username,
		Password:  ur.Password,
		Email:     ur.Email,
		FirstName: ur.FirstName,
		LastName:  ur.LastName,
	}
	if err := us.db.Create(&u).Error; err != nil {
		if isDuplicateError(err) {
			return u, utils.NewPublicError(err, fmt.Sprintf("Username %s is already taken", ur.Username), http.StatusBadRequest)
		}
		return u, utils.NewPrivateError(err)
	}
	return u, nil
}
