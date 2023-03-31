package database

import (
	"backendserver/utility"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connect() (err error) {
	db, err = sql.Open(utility.Config.Database.OpenConfig())
	return err
}
