package models

import (
	"database/sql"
	"time"
)

const dbTimeout = 3000 * time.Second

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}
