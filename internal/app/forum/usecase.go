package forum

import "github.com/efimovad/Forums.git/internal/models"

const (
	NOT_FOUND_ERR = "Can't find user with nickname "
	THREAD_CONFLICT = "Such thread already exists"
	FORUM_CONFLICT = "Such forum already exists"
	NOT_FOUND = "no such forum"
	THREAD_NOT_FOUND = "no such thread"
	VOTE_CONFLICT = "Such user already voted"
)

type Usecase interface {
	CreateForum(forum *models.Forum) (*models.Forum, error)
	GetForum(slug string) (*models.Forum, error)

	CreateThread(newThread *models.Thread) (*models.Thread, error)
	GetThreads(slug string, params *models.ListParameters) ([]*models.Thread, error)
	GetThread(currThread string) (*models.Thread, error)

	CreatePosts(currForum string, posts []*models.Post) error
	GetPosts(currThread string, params *models.ListParameters) ([]*models.Post, error)

	CreateVote(vote *models.Vote) (*models.Thread, error)
}
