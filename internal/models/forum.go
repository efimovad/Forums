package models

type Forum struct {
	ID		int64	`json:"-"`
	Posts	int64	`json:"posts"`
	Slug	string	`json:"slug"`
	Threads int64	`json:"threads"`
	Title	string	`json:"title"`
	User	string	`json:"user"`
}
