package models

import "time"

type Forum struct {
	ID		int64	`json:"-"`
	Slug	string	`json:"slug"`
	Title	string	`json:"title"`
	User	string	`json:"user"`
	Posts	int64	`json:"posts"`
	Threads	int64	`json:"threads"`
}

type Thread struct {
	ID		int64		`json:"id"`
	Forum	string		`json:"forum"`
	Author	string		`json:"author"`
	Created	time.Time	`json:"created"`
	Message	string		`json:"message"`
	Title	string		`json:"title"`
	Slug	string		`json:"slug"`
	Votes	int64		`json:"votes"`
}

type Post struct {
	ID			int64		`json:"id"`
	Author		string		`json:"author"`
	Created 	time.Time	`json:"created"`
	Forum		string		`json:"forum"`
	IsEdited	bool		`json:"isEdited"`
	Message		string		`json:"message"`
	Parent		int64		`json:"parent"`
	Thread		int64		`json:"thread"`
	Slug		string		`json:"slug"`
}