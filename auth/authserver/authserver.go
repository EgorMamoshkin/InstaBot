package authserver

import (
	"github.com/EgorMamoshkin/InstaBot/auth/handler"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"github.com/gorilla/mux"
	"net/http"
)

type AuthServer struct {
	handler *handler.ResponseHandler
}

func New(handler *handler.ResponseHandler) *AuthServer {
	return &AuthServer{handler: handler}
}

func (a *AuthServer) StartLS() error {
	r := mux.NewRouter()
	r.HandleFunc("/auth", a.handler.ServeHTTP)
	http.Handle("/", r)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return er.Wrap("authentication server broken", err)
	}

	return nil
}
