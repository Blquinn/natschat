package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"natschat/config"
	"natschat/models"
	"time"
)

type JWT struct {
	config *config.Config
}

func NewJWT(c *config.Config) *JWT {
	return &JWT{
		config: c,
	}
}

// AuthenticateUserJWT retrieves the user's information from the database
// and adds it to the gin context
func (j *JWT) AuthenticateUserJWT(c *gin.Context) {
	token, err := parseBearerToken(c)
	if err != nil {
		log.Debugln(err.Err.Error())
		c.JSON(err.Code, err.ResponseMap)
		c.Abort()
		return
	}
	user, ok := j.userIsAuthenticatedJWT(token)
	if !ok {
		c.JSON(401, authFailure)
		c.Abort()
		return
	}
	c.Set("user", user)
}

func (j *JWT) userIsAuthenticatedJWT(tokenString string) (models.JWTUser, bool) {
	var u models.JWTUser
	var err error
	if u, err = j.ParseAndValidateJWT(tokenString); err != nil {
		return u, false
	}
	return u, true
}

// ParseAndValidateJWT returns an error or a successfully parsed JWT Token
// func ParseAndValidateJWT(tokenString string) (*jwt.Token, error) {
func (j *JWT) ParseAndValidateJWT(tokenString string) (models.JWTUser, error) {

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	var user = models.JWTUser{}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(j.config.JWT.SecretKey), nil
	})

	if err != nil {
		return user, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return user, err
	}

	err = claims.Valid()
	if err != nil {
		return user, err
	}

	userID, okID := claims["user_id"].(float64)
	email, okEmail := claims["email"].(string)
	username, okUsername := claims["username"].(string)

	if !(okID && okEmail && okUsername) {
		return user, errors.New("failed to parse jwt claims")
	}
	user.ID = uint(userID)
	user.Email = email
	user.Username = username
	return user, nil
}

// creates a jwtString
func (j *JWT) CreateJWT(user models.User) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":    user.Email,
		"username": user.Username,
		"user_id":  uint32(user.ID),
		"exp":      time.Now().In(time.UTC).Add(time.Duration(j.config.JWT.ExpirySeconds) * time.Second).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(j.config.JWT.SecretKey))
	if err != nil {
		log.Errorf("An error occurred while creating JWT: %v", err)
		return tokenString, err
	}
	return tokenString, nil
}
