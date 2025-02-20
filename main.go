package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func LoadDataFromChannel(db *sql.DB, channel string) error {
	before, after, err := GetScoreIDRange(db)
	if err != nil {
		return err
	}
	if before != "" && after != "" {
		// Incremental load
		err = FetchFromDiscordAndPersist(nil, Options{Channel: channel, Before: before})
		err = FetchFromDiscordAndPersist(nil, Options{Channel: channel, After: after})
	} else {
		// Fetch all
		err = FetchFromDiscordAndPersist(nil, Options{Channel: channel})
	}
	if err != nil {
		return err
	}
	return nil
}

func MonitorChannel(db *sql.DB, channel string) error {
	err := LoadDataFromChannel(db, channel)
	if err != nil {
		return err
	}

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
	args := os.Args[1:]
	if len(args) == 0 || args[0][0] == '-' {
		args = append([]string{"serve"}, args...)
	}
	cmd := args[0]
	db, err := LoadDatabase()
	switch cmd {
	case "bot":
		ConnectToDiscord()
	case "monitor":
		cmd := flag.NewFlagSet("monitor", flag.ExitOnError)
		channel := cmd.String("channel", getDefaultChannel(), "Channel ID to monitor")
		cmd.Parse(args[1:])
		log.Printf("Monitoring channel <%s>", *channel)
		err = MonitorChannel(db, *channel)
	case "rescan":
		cmd := flag.NewFlagSet("rescan", flag.ExitOnError)
		channel := cmd.String("channel", getDefaultChannel(), "Channel ID to scan")
		cmd.Parse(args[1:])
		log.Printf("Full rescan of channel <%s>", *channel)
		err = FetchFromDiscordAndPersist(nil, Options{Channel: *channel})
	case "serve":
		cmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := cmd.String("port", "7654", "Port to run server")
		cmd.Parse(args[1:])
		addr := fmt.Sprintf(":%s", *port)
		log.Printf("Starting server on %s", addr)
		err = startServer(db, addr)
	case "stats":
		cmd := flag.NewFlagSet("serve", flag.ExitOnError)
		game := cmd.String("game", "Wordle", "Game to print stats")
		cmd.Parse(args[1:])
		stats, err := GetStats(db, *game)
		if err != nil {
			log.Fatal(err)
		}
		PrintStats(stats)
	default:
		help()
	}
	if err != nil {
		log.Fatal(err)
	}
}
