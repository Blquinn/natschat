package db

import "github.com/lib/pq"

func IsDuplicateError(err error) bool {
	if e, ok := err.(*pq.Error); ok && e.Code == "23505" {
		return true
	}
	return false
}
