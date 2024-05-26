package data

import "database/sql"

type UserRepo struct {
	DB *sql.DB
}
