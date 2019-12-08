package models

type User struct {
	ID			int64	`json:"-"`
	About 		string	`json:"about"`
	Email 		string	`json:"email"`
	FullName	string	`json:"fullname"`
	Nickname 	string	`json:"nickname"`
}
