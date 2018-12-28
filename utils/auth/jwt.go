package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"playground/natschat/models"
	"time"
)

var (
	jwtSecretKey = getJWTSecretKey()

	//jwtLeeway = 1
)

func getJWTSecretKey() []byte {
	key := os.Getenv("JWT_SECRET_KEY")
	if key != "" {
		return []byte(key)
	}
	log.Println("JWT_SECRET_KEY missing using default")
	return []byte("replace_me")
}

// AuthenticateUserJWT retrieves the user's information from the database
// and adds it to the gin context
func AuthenticateUserJWT(c *gin.Context) {
	token, err := parseBearerToken(c)
	if err != nil {
		log.Debugln(err.Err.Error())
		c.JSON(err.Code, err.ResponseMap)
		c.Abort()
		return
	}
	user, ok := userIsAuthenticatedJWT(token)
	if !ok {
		c.JSON(401, authFailure)
		c.Abort()
		return
	}
	c.Set("user", user)
}

func userIsAuthenticatedJWT(tokenString string) (models.User, bool) {
	user, err := ParseAndValidateJWT(tokenString)
	if err != nil {
		return models.User{}, false
	}
	return user, true
}

// ParseAndValidateJWT returns an error or a successfully parsed JWT Token
// func ParseAndValidateJWT(tokenString string) (*jwt.Token, error) {
func ParseAndValidateJWT(tokenString string) (models.User, error) {

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	var user = models.User{}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return jwtSecretKey, nil
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

	userID, okID := claims["user_id"].(string)
	email, okEmail := claims["email"].(string)
	username, okUsername := claims["username"].(string)

	if !(okID && okEmail && okUsername) {
		return user, errors.New("failed to parse jwt claims")
	}
	user.ID = userID
	user.Email = email
	user.Username = username
	return user, nil
}

// creates a jwtString
func CreateJWT(user *models.User) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":       user.Email,
		"username":    user.Username,
		"user_id":     user.ID,
		"exp":         time.Now().In(time.UTC).Add(15 * time.Minute).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}
