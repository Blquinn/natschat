package services

import (
	"github.com/jmoiron/sqlx"
	"playground/natschat/models"
)

type IUserService interface {
	GetUserByID(id string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)

}

var _ IUserService = &UserService{}

type UserService struct {
	db *sqlx.DB
}

func NewUserService(db *sqlx.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (us *UserService) GetUserByID(id string) (*models.User, error) {
	u := models.User{}
	err := us.db.Get(&u, `select * from users_user where id = $1`, id)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (us *UserService) GetUserByUsername(username string) (*models.User, error) {
	u := models.User{}
	err := us.db.Get(&u, `select * from users_user where username = $1`, username)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
