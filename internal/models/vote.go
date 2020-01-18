package models

type Vote struct {
	ID			int64	`json:"id"`
	Nickname	string	`json:"nickname"`
	Voice		int64	`json:"voice"`
	Thread		string	`json:"thread"`
}
