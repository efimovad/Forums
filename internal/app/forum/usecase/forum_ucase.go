package forum_ucase

import (
	"github.com/efimovad/Forums.git/internal/app/forum"
	"github.com/efimovad/Forums.git/internal/app/user"
	"github.com/efimovad/Forums.git/internal/models"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type ForumUcase struct {
	repository	forum.Repository
	userRep		user.Repository
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
		return errors.Wrap(errors.New(forum.NOT_FOUND), "forumRep.FindBySlug()")
	}

	created := time.Now().UTC()
	for _, elem := range posts {
		if elem.Slug != "" {
			if _, err := u.repository.FindPostBySlug(elem.Slug); err != nil {
				return errors.New("conflict")
			}
		}

		if elem.Parent != 0 {
			if _, err := u.repository.FindPost(elem.Parent); err != nil {
				return err
			}
		}

		_, err = u.userRep.FindByName(elem.Author)
		if err != nil {
			return errors.Wrap(errors.New(forum.NOT_FOUND_ERR), "userRep.FindByName()")
		}

		elem.Thread = t.ID
		elem.Forum = t.Forum
		elem.IsEdited = false
		elem.Created = created
	}

	err = u.repository.CreatePosts(posts)
	if err != nil {
		return errors.Wrap(err, "CreatePosts")
	}
	return nil
}

func (u *ForumUcase) CreateVote(vote *models.Vote) (*models.Thread, error) {
	thread, err := u.GetThread(vote.Thread)
	if err != nil {
		return nil, err
	}
	vote.Thread = thread.Slug

	v, err := u.repository.FindVote(vote.Thread, vote.Nickname)
	if err != nil {
		thread.Votes += vote.Voice
		if err := u.repository.CreateVote(vote); err != nil {
			return nil, err
		}
	} else {
		thread.Votes += vote.Voice - v.Voice
		v.Voice = vote.Voice

		if err := u.repository.UpdateVote(v); err != nil {
			return nil, err
		}
	}

	if err := u.repository.UpdateThread(thread); err != nil {
		return nil, err
	}

	return thread, nil
}

func (u *ForumUcase) GetThread(currThread string) (*models.Thread, error) {
	var threadFound bool
	var thread *models.Thread

	id, err := strconv.ParseInt(currThread, 10, 64)
	if err == nil {
		thread, err = u.repository.FindThread(id)
		if err == nil {
			threadFound = true
		}
	}

	if !threadFound {
		thread, err = u.repository.FindThreadBySlug(currThread)
		if err == nil {
			threadFound = true
		}
	}

	if !threadFound {
		return nil, errors.New(forum.THREAD_NOT_FOUND)
	}

	return thread, nil
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