package store

import (
	"database/sql"
	_ "github.com/lib/pq"
	)

func New(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	if err := createTables(db); err != nil {
		return nil, err
	}
	return db, nil
}

func createTables(db *sql.DB) error{
	userQuery := `CREATE TABLE IF NOT EXISTS users (
		email varchar unique not null,
		about varchar,
		fullname varchar,
		nickname varchar unique not null 
	);`
	if _, err := db.Exec(userQuery); err != nil {
		return err
	}
	return nil
}