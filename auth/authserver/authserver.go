package authserver

import (
	"fmt"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type AuthServer struct {
	router *mux.Router
}

func New(router *mux.Router) *AuthServer {
	return &AuthServer{router: router}
}

func (a *AuthServer) StartLS() error {
	r := mux.NewRouter()
	r.HandleFunc("/auth", RedirectToken)

	http.Handle("/", r)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return er.Wrap("authentication server broken", err)
	}

	return nil
}

func RedirectToken(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimRight(r.URL.Query().Get("code"), "#_")
	if code == "" {
		log.Println("there are no code")
	}

	link := fmt.Sprintf("https://telegram.me/share/url?url=/getaccess&text=%s", code)

	http.Redirect(w, r, link, http.StatusSeeOther)
}
