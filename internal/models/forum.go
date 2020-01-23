package models

import "time"

type Forum struct {
	ID		int64	`json:"-"`
	Slug	string	`json:"slug,omitempty"`
	Title	string	`json:"title,omitempty"`
	User	string	`json:"user,omitempty"`
	Posts	int64	`json:"posts,omitempty"`
	Threads	int64	`json:"threads,omitempty"`
}

type Thread struct {
	ID		int64		`json:"id,omitempty"`
	Forum	string		`json:"forum,omitempty"`
	Author	string		`json:"author,omitempty"`
	Created	time.Time	`json:"created"`
	Message	string		`json:"message,omitempty"`
	Title	string		`json:"title,omitempty"`
	Slug	string		`json:"slug,omitempty"`
	Votes	int64		`json:"votes,omitempty"`
}

type Post struct {
	ID			int64		`json:"id,omitempty"`
	Author		string		`json:"author,omitempty"`
	Created 	string		`json:"created,omitempty"`
	Forum		string		`json:"forum,omitempty"`
	IsEdited	bool		`json:"isEdited,omitempty"`
	Message		string		`json:"message,omitempty"`
	Parent		int64		`json:"parent"`
	Thread		int64		`json:"thread,omitempty"`
	Slug		string		`json:"slug,omitempty"`
	Path		[]int64		`json:"-"`
}