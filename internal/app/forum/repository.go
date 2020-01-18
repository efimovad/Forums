package forum

import "github.com/efimovad/Forums.git/internal/models"

type Repository interface {
	CreateForum(forum *models.Forum) error
	FindBySlug(slug string) (*models.Forum, error)
	FindByTitle(title string) (*models.Forum, error)

	CreateThread(thread *models.Thread) error
	GetThreads(slug string, params *models.ListParameters) ([]*models.Thread, error)
	FindThread(id int64) (*models.Thread, error)
	FindThreadBySlug(slug string) (*models.Thread, error)
	UpdateThread(thread *models.Thread) error

	CreatePosts(posts []*models.Post) error
	FindPost(id int64) (*models.Post, error)
	FindPostBySlug(slug string) (*models.Post, error)
	GetPosts(thread *models.Thread, params *models.ListParameters) ([]*models.Post, error)

	CreateVote(vote *models.Vote) error
	FindVote(thread string, nickname string) (*models.Vote, error)
	UpdateVote(vote *models.Vote) error
}