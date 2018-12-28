package auth

import (
	"testing"
)

func TestCraeteAndParseJWT(t *testing.T) {
	tokenString, err := createJWT()
	if err != nil {
		t.Error(err)
		return
	}
	user, err := ParseAndValidateJWT(tokenString)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(user)
}
