package general

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
)

type ctxKey int8

const (
	SessionName        = "user-session"
	CtxKeyUser              ctxKey = iota
)

func Error(w http.ResponseWriter, r *http.Request, code int, err error) {
	//log.Println(err)
	Respond(w, r, code, map[string]string{"message": errors.Cause(err).Error()})
}

func Respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
