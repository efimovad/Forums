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
    	id bigserial not null primary key,
		email varchar unique not null,
		about varchar,
		fullname varchar,
		nickname varchar unique not null 
	);`
	if _, err := db.Exec(userQuery); err != nil {
		return err
	}

	forumQuery := `CREATE TABLE IF NOT EXISTS forums (
    	id bigserial not null primary key,
		slug varchar unique not null,
		title varchar,
		"user" varchar not null 
	);`
	if _, err := db.Exec(forumQuery); err != nil {
		return err
	}

	threadQuery := `CREATE TABLE IF NOT EXISTS threads (
    	id bigserial not null primary key,
		forum varchar not null,
		author varchar,
		created timestamp,
		message varchar,
		title varchar,
		slug varchar,
		votes integer
	);`
	if _, err := db.Exec(threadQuery); err != nil {
		return err
	}

	postSequence := `CREATE SEQUENCE IF NOT EXISTS post_path 
						INCREMENT 1 
 						MINVALUE 1 
    					MAXVALUE 999999 
						START 1 
						CACHE 1;`
	if _, err := db.Exec(postSequence); err != nil {
		return err
	}

	postQuery := `CREATE TABLE IF NOT EXISTS posts (
    	id bigserial not null primary key,
    	path varchar, 
		author varchar,
		created timestamp,
		forum varchar not null,
		isEdited boolean,
		message varchar,
		parent integer,
		thread integer,
		slug varchar
	);`
	if _, err := db.Exec(postQuery); err != nil {
		return err
	}

	pathFunc := `CREATE OR REPLACE FUNCTION auto_id () returns varchar as $$
						select TO_CHAR(nextval('post_path'::regclass),'fm000000')
					$$ language sql `
	if _, err := db.Exec(pathFunc); err != nil {
		return err
	}

	/*postIndex := `CREATE INDEX path_post_idx ON post USING btree(path);`
	if _, err := db.Exec(postIndex); err != nil {
		return err
	}*/

	voteQuery := `CREATE TABLE IF NOT EXISTS votes (
    	id bigserial not null primary key,
		nickname varchar,
		vote integer,
		thread varchar
	);`
	if _, err := db.Exec(voteQuery); err != nil {
		return err
	}

	return nil
}