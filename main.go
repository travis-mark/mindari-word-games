package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func LoadDataFromChannel(db *sql.DB, channel string) error {
	before, after, err := GetScoreIDRange(db)
	if err != nil {
		return err
	}
	if before != "" && after != "" {
		// Incremental load
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel, Before: before})
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel, After: after})
	} else {
		// Fetch all
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel})
	}
	if err != nil {
		return err
	}
	return nil
}

func MonitorChannel(db *sql.DB, channel string) error {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := LoadDataFromChannel(db, channel)
			if err != nil {
				return err
			}
		}
	}
}

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
	cmd := ""
	if len(os.Args) < 2 {
		cmd = "channel"
	} else {
		cmd = os.Args[1]
	}
	switch cmd {
	case "channel":
		err = MonitorChannel(db, channel)
	case "rescan":
		err = FetchFromDiscordAndPersist(db, Options{Channel: channel})
	case "serve":
		err = startServer(db)
	case "stats":
		stats, err := GetStats(db, "Wordle")
		if err != nil {
			log.Fatal(err)
		}
		PrintStats(stats)
	default:
		log.Fatal(fmt.Sprintf("Command %s not found", cmd))
	}
	if err != nil {
		log.Fatal(err)
	}
}
