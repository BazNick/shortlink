package main

import "github.com/BazNick/shortlink/cmd/shortener"

func main() {
	if err := shortener.Run(); err != nil {
        panic(err)
    }
}