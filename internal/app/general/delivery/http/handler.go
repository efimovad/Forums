package general_handler

import (
	"github.com/efimovad/Forums.git/internal/app/general"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
)

type Handler struct {
	usecase			general.Usecase
	sessionStore	sessions.Store
}

func NewGeneralHandler(m *mux.Router, u general.Usecase, sessionStore sessions.Store) {
	handler := &Handler{
		usecase:		u,
		sessionStore:   sessionStore,
	}

	m.HandleFunc("/service/clear", handler.ClearService).Methods(http.MethodPost)
	m.HandleFunc("/service/status", handler.GetServiceStatus).Methods(http.MethodGet)
}

func (h *Handler) ClearService(w http.ResponseWriter, r *http.Request) {
	err := h.usecase.DropAll()
	if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}
	general.Respond(w, r, http.StatusOK, struct{}{})
}

func (h *Handler) GetServiceStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info, err := h.usecase.GetStatus()
	if err != nil {
		general.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	general.Respond(w, r, http.StatusOK, info)
}