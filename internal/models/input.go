package models

type ListParameters struct {
	Limit	int64	`json:"limit"`
	Since	string	`json:"since"`
	Desc	bool	`json:"desc"`
	Sort	string	`json:"sort"`
}