package config

import "github.com/jmoiron/sqlx"
import _ "github.com/lib/pq"

func GetDBConn() *sqlx.DB {
	return sqlx.MustConnect("postgres", "host=localhost port=5432 user=ben password=password dbname=chat sslmode=disable")
}
