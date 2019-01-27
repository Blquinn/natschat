// +build integration

package services

import (
	"natschat/models"
	"natschat/test"
	"testing"
)

func TestUserService_CreateUser(t *testing.T) {
	us := NewUserService(test.GetTestDB())
	ur := models.CreateUserRequest{
		Username:  "ben",
		Password:  "password",
		Email:     "ben@email.com",
		FirstName: "Ben",
		LastName:  "Quinn",
	}
	u, err := us.CreateUser(ur)
	if err != nil {
		t.Fatal(err)
	}

	if u.ID < 1 {
		t.Fatal("User not created successfully")
	}
}
