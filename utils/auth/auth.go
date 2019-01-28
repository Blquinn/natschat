package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Authentication middleware

var (
	// response json for auth
	authFailure      = map[string]string{"detail": "Authentication credentials were not provided."}
	permissionDenied = map[string]string{"detail": "You do not have permission to perform this action."}
	serverError      = map[string]string{"detail": "Internal server error."}
)

// authError has a response code and and error to be
// used as the response text
type authError struct {
	Code        int
	ResponseMap map[string]string
	Err         error
}

func GetUserOrPanic(c *gin.Context) JWTUser {
	user, exists := c.Get("user")
	if !exists {
		panic("`user` does not exist in gin context.")
	}
	userCasted, ok := user.(JWTUser)
	if !ok {
		panic("failed to cast context user to models.User")
	}
	return userCasted
}

func parseBearerToken(c *gin.Context) (string, *authError) {
	key := getHeader(c, "Authorization")
	if key == "" {
		log.Debugln("error: bearer token", key)
		return "", &authError{
			Code:        401,
			ResponseMap: authFailure,
			Err:         errors.New("error parsing bearer token"),
		}
	}

	karr := strings.Fields(key)
	if len(karr) != 2 {
		log.Debugln("error: bearer token", karr)
		return "", &authError{
			Code:        401,
			ResponseMap: authFailure,
			Err:         errors.New("error parsing bearer token"),
		}
	}

	return karr[1], nil
}

func getHeader(c *gin.Context, key string) string {
	if values := c.Request.Header[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}
