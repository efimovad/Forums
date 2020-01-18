package app

import (
	forum_handler "github.com/efimovad/Forums.git/internal/app/forum/delivery/http"
	forum_rep "github.com/efimovad/Forums.git/internal/app/forum/repository"
	forum_ucase "github.com/efimovad/Forums.git/internal/app/forum/usecase"
	general_handler "github.com/efimovad/Forums.git/internal/app/general/delivery/http"
	general_rep "github.com/efimovad/Forums.git/internal/app/general/repository"
	general_ucase "github.com/efimovad/Forums.git/internal/app/general/usecase"
	user_handler "github.com/efimovad/Forums.git/internal/app/user/delivery/http"
	user_rep "github.com/efimovad/Forums.git/internal/app/user/repository"
	user_ucase "github.com/efimovad/Forums.git/internal/app/user/usecase"
	"github.com/efimovad/Forums.git/internal/store"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type Server struct {
	config			*Config
	mux				*mux.Router
	sessionStore	sessions.Store
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) configure() error{
	myStore, err := store.New(s.config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "myStore.New()")
	}
	userRep := user_rep.NewUserRepository(myStore)
	generalRep := general_rep.NewGeneralRepository(myStore)
	forumRep := forum_rep.NewForumRepository(myStore)

	userUcase := user_ucase.NewUserUsecase(userRep)
	generalUcase := general_ucase.NewGeneralUsecase(generalRep)
	forumUcase := forum_ucase.NewForumUsecase(forumRep, userRep)

	user_handler.NewUserHandler(s.mux, userUcase, s.sessionStore)
	general_handler.NewGeneralHandler(s.mux, generalUcase, s.sessionStore)
	forum_handler.NewForumHandler(s.mux, forumUcase, s.sessionStore)

	return nil
}

func NewServer() *Server{
	config := NewConfig()
	return &Server{
		config:       	config,
		mux:          	mux.NewRouter(),
		sessionStore:	sessions.NewCookieStore([]byte(config.SessionKey)),
	}
}

func Start() error {
	server := NewServer()
	if err := server.configure(); err != nil {
		return errors.Wrap(err, "server.configure()")
	}
	log.Println("running server on ", server.config.BindAddr, "...")
	return http.ListenAndServe(server.config.BindAddr, server)
}
