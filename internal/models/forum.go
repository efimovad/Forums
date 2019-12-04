package models

type Forum struct {
	Slug string `json:"slug"`
	Title string `json:"title"`
	User string `json:"user"`
	Posts int64 `json:"posts"`
	Threads int64 `json:"threads"`
}
