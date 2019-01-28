package apierrs

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

var (
	errInternalServerError = map[string]string{"Detail": "Internal server error occurred."}
)

type APIError struct {
	InternalError error
	Message       string
	// IsPublic designates whether this error should be displayed to the user.
	IsPublic     bool
	ResponseCode int
}

func (e *APIError) ResponseJSON() map[string]string {
	if e.IsPublic {
		return map[string]string{
			"Detail": e.Message,
		}
	}
	return errInternalServerError
}

func (e *APIError) HandleResponse(c *gin.Context) {
	if !e.IsPublic {
		c.JSON(http.StatusInternalServerError, e.ResponseJSON())
		return
	}

	var code int
	if e.ResponseCode == 0 {
		code = http.StatusInternalServerError
	} else {
		code = e.ResponseCode
	}

	c.JSON(code, e.ResponseJSON())
}

func HandleValidationError(c *gin.Context, err error) {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		c.JSON(http.StatusInternalServerError, errInternalServerError)
		c.Abort()
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"Detail": "Validation error occurred.",
		"Errors": FormatValidationErrors(ve),
	})
	c.Abort()
}

type ValidationError struct {
	Field   string
	Message string
}

func FormatValidationErrors(e validator.ValidationErrors) []ValidationError {
	ves := make([]ValidationError, len(e))
	i := 0
	for k, v := range e {
		ves[i] = ValidationError{
			Field:   k,
			Message: fmt.Sprintf("Validation on field `%s` failed on tag `%s`", v.Field, v.Tag),
		}
		i++
	}
	return ves
}

func (e *APIError) Error() string {
	return e.InternalError.Error()
}

var _ error = &APIError{}

func NewPublicError(err error, message string, responseCode int) *APIError {
	return &APIError{
		InternalError: err,
		Message:       message,
		IsPublic:      true,
		ResponseCode:  responseCode,
	}
}

func NewPrivateError(err error) *APIError {
	return &APIError{
		InternalError: err,
		Message:       err.Error(),
		IsPublic:      false,
	}
}
