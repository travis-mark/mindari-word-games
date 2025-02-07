package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env to get secrets
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	channel := os.Getenv("CHANNEL")
	err = fetchFromDiscord(Options{Channel: channel})
	if err != nil {
		log.Fatal(err)
	}
}
