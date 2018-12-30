package utils

type APIError struct {
	InternalError error
	Message       string
	// IsPublic designates whether this error should be displayed to the user.
	IsPublic      bool
}

func (e *APIError) ResponseJSON() map[string]string {
	if e.IsPublic {
		return map[string]string{
			"detail": e.Message,
		}
	}
	return map[string]string{
		"detail": "Internal server error occurred.",
	}
}

func (e *APIError) Error() string {
	return e.InternalError.Error()
}

var _ error = &APIError{}

func NewPubicError(err error, message string) *APIError {
	return &APIError{
		InternalError: err,
		Message: message,
		IsPublic: true,
	}
}

func NewPrivateError(err error) *APIError {
	return &APIError{
		InternalError: err,
		Message: err.Error(),
		IsPublic: false,
	}
}
