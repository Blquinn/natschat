package users

import (
	"gopkg.in/go-playground/validator.v8"
	"testing"
)

func TestCreateUserRequestValidation(t *testing.T) {
	r := CreateUserRequest{}

	validate := validator.New(&validator.Config{TagName: "validate"})

	err := validate.Struct(&r)
	if err == nil {
		t.Error("Failed to validate struct")
		return
	}
}
