package models

type ServiceInfo struct {
	Forum	int64	`json:"forum"`
	Post	int64	`json:"post"`
	Thread	int64	`json:"thread"`
	User	int64	`json:"user"`
}

type Combine struct {
	Post *Post `json:"post"`
	Forum *Forum `json:"forum"`
	Thread *Thread `json:"thread"`
	Author *User `json:"author"`
}