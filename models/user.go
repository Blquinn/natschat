package models

import "time"

type User struct {
	ID string `db:"id"` // UUID
	Username string `db:"username"`
	Password string `db:"password"`
	Email string `db:"email"`
	FirstName string `db:"first_name"`
	LastName string `db:"last_name"`
	InsertedAt time.Time `db:"inserted_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
