package main

import (
	"net/http"

	"github.com/BazNick/shortlink/internal/app/api"
)


func main() {
	mux := http.NewServeMux()

	mux.HandleFunc(`/`, api.AddLink)
	mux.HandleFunc(`/{id}`, api.GetLink)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
