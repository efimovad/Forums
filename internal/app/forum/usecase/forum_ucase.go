package forum_ucase

import (
	"github.com/efimovad/Forums.git/internal/app/forum"
	"github.com/efimovad/Forums.git/internal/app/user"
	"github.com/efimovad/Forums.git/internal/models"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
)

type ForumUcase struct {
	repository	forum.Repository
	userRep		user.Repository
	mux			sync.Mutex
}

func NewForumUsecase(r forum.Repository, ur user.Repository) forum.Usecase {
	return &ForumUcase{
		repository: r,
		userRep:	ur,
	}
}

func (u *ForumUcase) CreateForum(newForum *models.Forum) (*models.Forum, error) {
	f, err := u.repository.FindBySlug(newForum.Slug)
	if err == nil {
		return f, errors.Wrap(errors.New(forum.FORUM_CONFLICT), "forumRep.FindBySlug()")
	}

	us, err := u.userRep.FindByName(newForum.User)
	if err != nil {
		return f, errors.Wrap(errors.New(forum.NOT_FOUND_ERR), "userRep.FindByName()")
	}

	newForum.User = us.Nickname

	return nil, u.repository.CreateForum(newForum)
}

func (u *ForumUcase) CreateThread(newThread *models.Thread) (*models.Thread, error) {
	if newThread.Slug != "" {
		t, err := u.repository.FindThreadBySlug(newThread.Slug)
		if newThread.Slug != "" && err == nil {
			return t, errors.Wrap(errors.New(forum.THREAD_CONFLICT), "forumRep.FindBySlug()")
		}
	}

	f, err := u.repository.FindBySlug(newThread.Forum)
	if err != nil {
		return nil, errors.Wrap(errors.New(forum.NOT_FOUND), "forumRep.FindBySlug()")
	}

	newThread.Forum = f.Slug

	us, err := u.userRep.FindByName(newThread.Author)
	if err != nil {
		return nil, errors.Wrap(errors.New(forum.NOT_FOUND_ERR), "userRep.FindByName()")
	}

	newThread.Author = us.Nickname

	return nil, u.repository.CreateThread(newThread)
}

func (u *ForumUcase) GetForum(slug string) (*models.Forum, error) {
	f, err := u.repository.FindBySlug(slug)
	if err != nil {
		return nil, errors.Wrap(errors.New(forum.NOT_FOUND), "forumRep.FindBySlug()")
	}
	return f, nil
}

func (u *ForumUcase) GetThreads(slug string, params *models.ListParameters) ([]*models.Thread, error) {
	_, err := u.repository.FindBySlug(slug)
	if err != nil {
		return nil, errors.Wrap(errors.New(forum.NOT_FOUND), "forumRep.FindBySlug()")
	}

	list, err := u.repository.GetThreads(slug, params)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (u *ForumUcase) CreatePosts(currForum string, posts []*models.Post) error {
	t, err := u.GetThread(currForum)
	if err != nil {
		return err
	}

	/*for _, elem := range posts {
		var parent *models.Post

		if elem.Parent == 0 {
			continue
		}

		parent, err = u.repository.FindPost(elem.Parent)
		if err != nil {
			return errors.New(forum.PARENT_POST_CONFLICT)
		}

		if parent != nil && parent.Thread != t.ID {
			return errors.New(forum.PARENT_POST_CONFLICT)
		}

		/*_, err = u.userRep.FindByName(elem.Author)
		if err != nil {
			return errors.Wrap(errors.New(forum.NOT_FOUND_ERR + elem.Author), "userRep.FindByName()")
		}
	}*/

	if len(posts) == 0 {
		return nil
	}

	err = u.repository.CreatePosts(posts, t)
	if err != nil {
		if strings.Contains(err.Error(), "posts_author_fkey") {
			return errors.Wrap(errors.New(forum.NOT_FOUND_ERR), "userRep.FindByName()")
		}
		return errors.Wrap(err, "CreatePosts")
	}
	return nil
}

func (u *ForumUcase) CreateVote(vote *models.Vote) (*models.Thread, error) {
	if vote.Voice != 1 && vote.Voice != -1 {
		return nil, errors.New(forum.WRONG_INPUT)
	}

	thread, err := u.GetThread(vote.Thread)
	if err != nil {
		return nil, err
	}

	if _, err = u.repository.FindUser(vote.Nickname); err != nil {
		return nil, errors.New("Can't find user by nickname: " + vote.Nickname)
	}

	votesNum, err := u.repository.CreateVote(vote, thread)
	if err != nil {
		return nil, err
	}

	thread.Votes = votesNum

	return thread, nil
}

func (u *ForumUcase) GetThread(currThread string) (*models.Thread, error) {
	var thread *models.Thread

	id, err := strconv.ParseInt(currThread, 10, 64)
	if err == nil {
		thread, err = u.repository.FindThread(id)
		if err != nil {
			return nil, errors.New("Can't find post thread by id: " + strconv.FormatInt(id, 10))
		}
		return thread, nil
	}

	thread, err = u.repository.FindThreadBySlug(currThread)
	if err != nil {
		return nil, errors.New("Can't find post thread by slug: " + currThread)
	}
	return thread, nil
}

func (u *ForumUcase) UpdateThread(currThread string, thread *models.Thread) (*models.Thread, error) {
	var exThread *models.Thread
	id, err := strconv.ParseInt(currThread, 10, 64)
	if err == nil {
		exThread, err = u.repository.FindThread(id)
	} else {
		exThread, err = u.repository.FindThreadBySlug(currThread)
	}

	if err != nil {
		return nil, errors.New(forum.THREAD_NOT_FOUND)
	}

	if thread.Title == "" && thread.Message == "" {
		return exThread, nil
	}

	if thread.Message != "" {
		exThread.Message = thread.Message
	}

	if thread.Title != "" {
		exThread.Title = thread.Title
	}

	if err = u.repository.UpdateThread(exThread); err != nil {
		return nil, errors.Wrap(err, "repository.UpdateThread")
	}

	return exThread, nil
}

func (u *ForumUcase) GetPosts(currThread string, params *models.ListParameters) ([]*models.Post, error){
	t, err := u.GetThread(currThread)
	if err != nil {
		return nil, errors.New(forum.THREAD_NOT_FOUND)
	}

	posts, err := u.repository.GetPosts(t, params)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (u *ForumUcase) FindPost(id int64) (*models.Post, error) {
	post, err := u.repository.FindPost(id)
	if err != nil {
		return nil, errors.New(forum.POST_NOT_FOUND)
	}
	return post, nil
}

func (u *ForumUcase) FindPostDetail(id int64, related string) (*models.Combine, error) {
	res := new(models.Combine)

	post, err := u.repository.FindPost(id)
	if err != nil {
		return nil, errors.New(forum.POST_NOT_FOUND)
	}

	res.Post = post

	if strings.Contains(related, "forum") {
		postForum, err := u.repository.FindBySlug(post.Forum)
		if err != nil {
			return nil, errors.New(forum.NOT_FOUND)
		}
		res.Forum = postForum
	}

	if strings.Contains(related, "thread") {
		postThread, err := u.repository.FindThread(post.Thread)
		if err != nil {
			return nil, errors.New(forum.NOT_FOUND)
		}
		res.Thread = postThread
	}

	if strings.Contains(related, "user") {
		postAuthor, err := u.repository.FindUser(post.Author)
		if err != nil {
			return nil, errors.New(forum.NOT_FOUND)
		}
		res.Author = postAuthor
	}

	return res, nil
}

func (u *ForumUcase) UpdatePost(post *models.Post) (*models.Post, error) {
	currPost, err := u.FindPost(post.ID)
	if err != nil {
		return nil, err
	}

	if post.Message == "" || post.Message == currPost.Message {
		return currPost, nil
	}

	currPost.Message = post.Message
	currPost.IsEdited = true

	if err = u.repository.UpdatePost(currPost); err != nil {
		return nil, err
	}

	return currPost, nil
}

func (u *ForumUcase) GetUsers(slug string, params models.ListParameters) ([]*models.User, error) {
	currForum, err := u.repository.FindBySlug(slug)
	if err != nil {
		return nil, errors.New(forum.NOT_FOUND)
	}

	users, err := u.repository.GetUsers(currForum.ID, params)
	if err != nil {
		return nil, errors.Wrap(err, "repository.GetUsers()")
	}
	return users, nil
}