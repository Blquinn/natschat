package users

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"natschat/models"
	"natschat/utils/apierrs"
	"natschat/utils/auth"
	"natschat/utils/db"
	"net/http"
)

type Service struct {
	jwt *auth.JWT
	db  *gorm.DB
}

func NewService(db *gorm.DB, jwt *auth.JWT) *Service {
	return &Service{
		db:  db,
		jwt: jwt,
	}
}

func (us *Service) LoginUser(r LoginRequest) (string, *apierrs.APIError) {
	var err error
	var user models.User
	if err := us.db.First(&user, "username = ?", r.Username).Error; err != nil {
		return "", apierrs.NewPublicError(err, "Incorrect username or password", http.StatusUnauthorized)
	}

	if user.Password != r.Password {
		return "", apierrs.NewPublicError(err, "Incorrect username or password", http.StatusUnauthorized)
	}

	var jwt string
	if jwt, err = us.jwt.CreateJWT(user.Email, user.Username, user.PublicID, user.ID); err != nil {
		return "", apierrs.NewPrivateError(err)
	}

	return jwt, nil
}

func (us *Service) CreateUser(ur CreateUserRequest) (UserDTO, *apierrs.APIError) {
	u := models.NewUser(ur.Username, ur.Password, ur.Email, ur.FirstName, ur.LastName)
	if err := us.db.Create(&u).Error; err != nil {
		if db.IsDuplicateError(err) {
			return UserDTO{}, apierrs.NewPublicError(err, fmt.Sprintf("Username %s is already taken", ur.Username), http.StatusBadRequest)
		}
		return UserDTO{}, apierrs.NewPrivateError(err)
	}
	return NewUserDTO(u.PublicID, u.Username, u.Email, u.FirstName, u.LastName), nil
}
