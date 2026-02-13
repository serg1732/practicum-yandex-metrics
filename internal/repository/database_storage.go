package repository

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

func BuildDataBase(DSL string) (*DataBase, error) {
	if DSL == "" {
		return nil, errors.New("DSL required")
	}
	db, err := sql.Open("postgres", DSL)
	if err != nil {
		return nil, err
	}
	if errPing := db.Ping(); errPing != nil {
		return nil, errPing
	}
	return &DataBase{db}, nil
}

type DataBase struct {
	db *sql.DB
}

func (db *DataBase) Ping() error {
	return db.db.Ping()
}
