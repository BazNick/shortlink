package shortener

import (
	"net/http"

	"github.com/BazNick/shortlink/internal/app/api"
)


func Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc(`/`, api.AddLink)
	mux.HandleFunc(`/{id}`, api.GetLink)

	return http.ListenAndServe(`:8080`, mux)
}
