package main

import (
	"log"
	"os"
)

func main() {
	// Это должно стриггерить анализатор
	os.Exit(1)
}

func otherFunc() {
	// Это тоже должно стриггерить анализатор, тк это в пакете main
	log.Fatal("error")
}
