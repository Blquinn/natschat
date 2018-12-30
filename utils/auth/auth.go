package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	log "github.com/sirupsen/logrus"
	"natschat/models"
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

func GetUserOrPanic(c *gin.Context) models.User {
	user, exists := c.Get("user")
	if !exists {
		panic("`user` does not exist in gin context.")
	}
	userCasted, ok := user.(models.User)
	if !ok {
		panic("failed to cast context user to models.User")
	}
	return userCasted
}

// UserVerified checks that a user has a valid auth token and
// that their account has been marked as verified
func UserVerified(c *gin.Context) {
	u, exists := c.Get("user")
	if !exists {
		log.Println("user struct not available in gin context")
		c.JSON(500, serverError)
		c.Abort()
		return
	}

	_, ok := u.(models.User)

	if !ok {
		c.JSON(500, serverError)
		c.Abort()
		log.Errorln(stacktrace.NewError("failed to cast user in gin context"))
		return
	}

	c.Next()
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
