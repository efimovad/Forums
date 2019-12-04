package models

type ServiceInfo struct {
	Forum	int64	`json:"forum"`
	Post	int64	`json:"post"`
	Thread	int64	`json:"thread"`
	User	int64	`json:"user"`
}
