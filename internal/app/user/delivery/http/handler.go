package user_handler

import (
	"encoding/json"
	"github.com/efimovad/Forums.git/internal/app/general"
	"github.com/efimovad/Forums.git/internal/app/user"
	"github.com/efimovad/Forums.git/internal/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

type Handler struct {
	usecase			user.Usecase
	sessionStore	sessions.Store
}

func NewUserHandler(m *mux.Router, u user.Usecase, sessionStore sessions.Store) {
	handler := &Handler{
		usecase:		u,
		sessionStore:   sessionStore,
	}

	m.HandleFunc("/", handler.MainHandler)
	m.HandleFunc("/api/user/{nickname}/create", handler.CreateUser).Methods(http.MethodPost)
	m.HandleFunc("/api/user/{nickname}/profile", handler.GetUser).Methods(http.MethodGet)
	m.HandleFunc("/api/user/{nickname}/profile", handler.EditUser).Methods(http.MethodPost)
}

func (h *Handler) MainHandler(w http.ResponseWriter, r *http.Request) {
	general.Respond(w, r, http.StatusOK, "HELLO FROM SERVER")
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "UserHandler.CreateUser<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	newUser := new(models.User)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(newUser)
	if err != nil {
		err = errors.Wrapf(err, "UserHandler.CreateUser<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	vars := mux.Vars(r)
	name := vars["nickname"]
	newUser.Nickname = name

	users, err := h.usecase.Create(newUser)
	if err != nil {
		general.Respond(w, r, http.StatusConflict, &users)
		return
	}
	general.Respond(w, r, http.StatusCreated, newUser)
}

func (h * Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["nickname"]

	currUser, err := h.usecase.FindByName(name)
	if err != nil {
		general.Error(w, r, http.StatusNotFound, errors.New("Can't find user with nickname " + name))
		return
	}

	general.Respond(w, r, http.StatusOK, currUser)
}


func (h * Handler) EditUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["nickname"]

	defer func() {
		if err := r.Body.Close(); err != nil {
			err = errors.Wrapf(err, "UserHandler.CreateUser<-r.Body.Close()")
			general.Error(w, r, http.StatusInternalServerError, err)
		}
	}()

	newUser := new(models.User)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(newUser)
	if err != nil {
		err = errors.Wrapf(err, "UserHandler.CreateUser<-Decode()")
		general.Error(w, r, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Edit(name, newUser); err != nil {
		if strings.Contains(err.Error(), user.NOT_FOUND_ERR) {
			general.Error(w, r, http.StatusNotFound, err)
		} else if strings.Contains(err.Error(), user.NICKNAME_CONFLICT) ||
			strings.Contains(err.Error(), user.EMAIL_CONFLICT) {
			general.Error(w,r, http.StatusConflict, err)
		} else {
			general.Error(w,r, http.StatusInternalServerError, err)
		}
		return
	}

	general.Respond(w, r, http.StatusOK, newUser)
}

