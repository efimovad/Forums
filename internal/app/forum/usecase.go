package forum

import "github.com/efimovad/Forums.git/internal/models"

const (
	NOT_FOUND = "Can't find such forum"
	NOT_FOUND_ERR = "Can't find post author by nickname: "
	THREAD_NOT_FOUND = "Can't find such thread"
	PARENT_POST_CONFLICT = "Parent post was created in another thread"
	THREAD_CONFLICT = "Such thread already exists"
	FORUM_CONFLICT = "Such forum already exists"
	VOTE_CONFLICT = "Such user already voted"
	POST_NOT_FOUND = "Can't find such post"
	USER_NOT_FOUND = "Can't find user by nickname: "
	WRONG_INPUT = "Wrong input"
)

type Usecase interface {
	CreateForum(forum *models.Forum) (*models.Forum, error)
	GetForum(slug string) (*models.Forum, error)
	GetUsers(slug string, params models.ListParameters) ([]*models.User, error)

	CreateThread(newThread *models.Thread) (*models.Thread, error)
	GetThreads(slug string, params *models.ListParameters) ([]*models.Thread, error)
	GetThread(currThread string) (*models.Thread, error)
	UpdateThread(currThread string, thread *models.Thread) (*models.Thread, error)

	CreatePosts(currForum string, posts []*models.Post) error
	GetPosts(currThread string, params *models.ListParameters) ([]*models.Post, error)
	FindPost(id int64) (*models.Post, error)
	FindPostDetail(id int64, related string) (*models.Combine, error)
	UpdatePost(post *models.Post) (*models.Post, error)

	CreateVote(vote *models.Vote) (*models.Thread, error)
}
