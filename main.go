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
	db, err := LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	before, after, err := GetScoreIDRange(db)
	if before != "" && after != "" {
		// Incremental load
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel, Before: before})
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel, After: after})
	} else {
		// Fetch all
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel})
	}
	if err != nil {
		log.Fatal(err)
	}
}
