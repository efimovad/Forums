package forum_handler

import (
	"encoding/json"
	"github.com/efimovad/Forums.git/internal/app/forum"
	"github.com/efimovad/Forums.git/internal/app/general"
	"github.com/efimovad/Forums.git/internal/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	usecase			forum.Usecase
	sessionStore	sessions.Store
}

func NewForumHandler(m *mux.Router, u forum.Usecase, sessionStore sessions.Store) {
	handler := &Handler{
		usecase:		u,
		sessionStore:   sessionStore,
	}

	m.HandleFunc("/api/forum/create", handler.CreateForum).Methods(http.MethodPost)
	m.HandleFunc("/api/forum/{slug}/create", handler.CreateThread).Methods(http.MethodPost)
	m.HandleFunc("/api/forum/{slug}/details", handler.GetForum).Methods(http.MethodGet)
	m.HandleFunc("/api/forum/{slug}/threads", handler.GetThreads).Methods(http.MethodGet)
	m.HandleFunc("/api/forum/{slug}/users", handler.GetUsers).Methods(http.MethodGet)

	m.HandleFunc("/api/thread/{slug_or_id}/create", handler.CreatePost).Methods(http.MethodPost)
	m.HandleFunc("/api/thread/{slug_or_id}/vote", handler.VoteThread).Methods(http.MethodPost)
	m.HandleFunc("/api/thread/{slug_or_id}/details", handler.GetThread).Methods(http.MethodGet)
	m.HandleFunc("/api/thread/{slug_or_id}/details", handler.UpdateThread).Methods(http.MethodPost)
	m.HandleFunc("/api/thread/{slug_or_id}/posts", handler.GetPosts).Methods(http.MethodGet)

	m.HandleFunc("/api/post/{id}/details", handler.GetPost).Methods(http.MethodGet)
	m.HandleFunc("/api/post/{id}/details", handler.UpdatePost).Methods(http.MethodPost)
}

func (h *Handler) CreateForum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.CreateForum<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	newForum := new(models.Forum)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(newForum)
	if err != nil {
		err = errors.Wrapf(err, "ForumHandler.CreateForum<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	exForum, err := h.usecase.CreateForum(newForum)
	if err != nil {
		if strings.Contains(err.Error(), forum.FORUM_CONFLICT) {
			general.Respond(w, r, http.StatusConflict, exForum)
		} else if strings.Contains(err.Error(), forum.NOT_FOUND_ERR) {
			general.Error(w,r, http.StatusNotFound, errors.New(forum.NOT_FOUND_ERR + newForum.User))
		} else {
			general.Error(w,r, http.StatusInternalServerError, err)
		}
		return
	}
	general.Respond(w, r, http.StatusCreated, newForum)
}

func (h *Handler) CreateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.CreateForum<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	vars := mux.Vars(r)
	slug := vars["slug"]

	newThread := new(models.Thread)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(newThread)
	if err != nil {
		err = errors.Wrapf(err, "ForumHandler.CreateForum<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	newThread.Forum = slug
	newThread.Created = newThread.Created.UTC()

	exThread, err := h.usecase.CreateThread(newThread)
	if err != nil {
		if strings.Contains(err.Error(), forum.NOT_FOUND) {
			general.Respond(w, r, http.StatusNotFound, errors.New("Can't find forum by slug: " + newThread.Forum))
		} else if strings.Contains(err.Error(), forum.THREAD_CONFLICT) {
			general.Respond(w,r, http.StatusConflict, exThread)
		} else if strings.Contains(err.Error(), forum.NOT_FOUND_ERR) {
			general.Error(w,r, http.StatusNotFound, errors.New(forum.NOT_FOUND_ERR + newThread.Author))
		} else {
			general.Error(w,r, http.StatusInternalServerError, err)
		}
		return
	}
	general.Respond(w, r, http.StatusCreated, newThread)
}

func (h *Handler) GetForum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]

	f, err := h.usecase.GetForum(slug)
	if err != nil {
		general.Error(w, r, http.StatusNotFound, errors.New("Can't find forum by slug: " + slug))
		return
	}
	general.Respond(w, r, http.StatusOK, f)
}

func (h *Handler) GetThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.GetThreads<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	vars := mux.Vars(r)
	slug := vars["slug"]

	params := new(models.ListParameters)
	str := r.URL.Query().Get("limit")
	if str == "" {
		params.Limit = 0
	} else {
		limit, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			general.Error(w, r, http.StatusBadRequest, err)
			return
		}
		params.Limit = limit
	}

	str = r.URL.Query().Get("desc")
	if str == "" {
		params.Desc = false
	} else {
		desc, err := strconv.ParseBool(str)
		if err != nil {
			general.Error(w, r, http.StatusBadRequest, err)
			return
		}
		params.Desc = desc
	}

	params.Since = r.URL.Query().Get("since")

	list, err := h.usecase.GetThreads(slug, params)
	if err != nil {
		general.Error(w, r, http.StatusNotFound, errors.New("Can't find forum by slug: " + slug))
		return
	}

	if len(list) == 0 {
		general.Respond(w, r, http.StatusOK,  []string{})
		return
	}
	general.Respond(w, r, http.StatusOK, list)
}

func (h *Handler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.CreateForum<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	var thread *models.Thread
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&thread)
	if err != nil {
		err = errors.Wrapf(err, "ForumHandler.UpdateThread()<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	vars := mux.Vars(r)
	slug := vars["slug_or_id"]

	res, err := h.usecase.UpdateThread(slug, thread)
	if err != nil && strings.Contains(err.Error(), "Can't find") {
		general.Error(w, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	general.Respond(w, r, http.StatusOK, res)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.CreateForum<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()


	vars := mux.Vars(r)
	slugOrID := vars["slug_or_id"]

	var list []*models.Post
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&list)
	if err != nil {
		err = errors.Wrapf(err, "ForumHandler.CreateForum<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	err = h.usecase.CreatePosts(slugOrID, list)
	if err != nil && strings.Contains(err.Error(), forum.PARENT_POST_CONFLICT) {
		general.Error(w, r, http.StatusConflict, err)
		return
	} else if err != nil && strings.Contains(err.Error(), "Can't find post thread by") {
		general.Error(w, r, http.StatusNotFound, err)
		return
	} else if err != nil && strings.Contains(err.Error(), forum.NOT_FOUND_ERR) {
		general.Error(w, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	general.Respond(w, r, http.StatusCreated, list)
}

func (h *Handler) VoteThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.CreateForum<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	vars := mux.Vars(r)
	slugOrID := vars["slug_or_id"]

	vote := new(models.Vote)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(vote)
	if err != nil {
		err = errors.Wrapf(err, "ForumHandler.CreateForum<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}
	vote.Thread = slugOrID

	thread, err := h.usecase.CreateVote(vote)
	if err != nil && strings.Contains(err.Error(), "Can't find") {
		general.Error(w, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	general.Respond(w, r, http.StatusOK, thread)
}

func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug_or_id"]

	t, err := h.usecase.GetThread(slug)
	if err != nil {
		general.Error(w, r, http.StatusNotFound, err)
		return
	}
	general.Respond(w, r, http.StatusOK, t)
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	currForum := vars["slug_or_id"]

	params := new(models.ListParameters)
	str := r.URL.Query().Get("limit")
	if str == "" {
		params.Limit = 0
	} else {
		limit, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			general.Error(w, r, http.StatusBadRequest, err)
			return
		}
		params.Limit = limit
	}

	str = r.URL.Query().Get("desc")
	if str == "" {
		params.Desc = false
	} else {
		desc, err := strconv.ParseBool(str)
		if err != nil {
			general.Error(w, r, http.StatusBadRequest, err)
			return
		}
		params.Desc = desc
	}

	params.Since = r.URL.Query().Get("since")
	params.Sort = r.URL.Query().Get("sort")
	if params.Sort == "" {
		params.Sort = "flat"
	}

	list, err := h.usecase.GetPosts(currForum, params)
	if err != nil && strings.Contains(err.Error(), "Can't find") {
		general.Error(w, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	if len(list) == 0 {
		general.Respond(w, r, http.StatusOK,  []string{})
		return
	}
	general.Respond(w, r, http.StatusOK, list)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	currForum := vars["slug"]

	var params models.ListParameters
	str := r.URL.Query().Get("limit")
	if str == "" {
		params.Limit = 0
	} else {
		limit, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			general.Error(w, r, http.StatusBadRequest, err)
			return
		}
		params.Limit = limit
	}

	str = r.URL.Query().Get("desc")
	if str == "" {
		params.Desc = false
	} else {
		desc, err := strconv.ParseBool(str)
		if err != nil {
			general.Error(w, r, http.StatusBadRequest, err)
			return
		}
		params.Desc = desc
	}

	params.Since = r.URL.Query().Get("since")

	list, err := h.usecase.GetUsers(currForum, params)
	if err != nil && err.Error() == forum.NOT_FOUND {
		general.Error(w, r, http.StatusNotFound, errors.New("Can't find forum by slug: " + currForum))
		return
	} else if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	if len(list) == 0 {
		general.Respond(w, r, http.StatusOK,  []string{})
		return
	}
	general.Respond(w, r, http.StatusOK, list)
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	currPost := vars["id"]

	id, err := strconv.ParseInt(currPost, 10, 64)
	if err != nil {
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	related := r.URL.Query().Get("related")

	post, err := h.usecase.FindPostDetail(id, related)
	if err != nil && err.Error() == forum.POST_NOT_FOUND {
		general.Error(w, r, http.StatusNotFound, errors.New("Can't find post by id: " + strconv.FormatInt(id, 10)))
		return
	} else if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	general.Respond(w, r, http.StatusOK, post)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	currPost := vars["id"]

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "ForumHandler.CreateForum<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	post := new(models.Post)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(post)
	if err != nil {
		err = errors.Wrapf(err, "ForumHandler.UpdatePost<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	id, err := strconv.ParseInt(currPost, 10, 64)
	if err != nil {
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	post.ID = id
	res, err := h.usecase.UpdatePost(post)
	if err != nil && err.Error() == forum.POST_NOT_FOUND {
		general.Error(w, r, http.StatusNotFound, errors.New("Can't find post by id: " + strconv.FormatInt(id, 10)))
		return
	} else if err != nil {
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	general.Respond(w, r, http.StatusOK, res)
}