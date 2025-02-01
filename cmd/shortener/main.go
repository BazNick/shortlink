package main

import (
	"net/http"

	"github.com/BazNick/shortlink/internal/app/handlers"
)


func main() {
	mux := http.NewServeMux()

	mux.HandleFunc(`/`, handlers.AddLink)
	mux.HandleFunc(`/{id}`, handlers.GetLink)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
