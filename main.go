package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0][0] == '-' {
		args = append([]string{"help"}, args...)
	}
	cmd := args[0]
	var err error
	switch cmd {
	case "monitor":
		err = MonitorDiscord()
	case "rescan":
		cmd := flag.NewFlagSet("rescan", flag.ExitOnError)
		channel := cmd.String("channel", "", "Channel ID to scan")
		cmd.Parse(args[1:])
		if *channel == "" {
			cmd.Usage()
			os.Exit(1)
		}
		err = ScanChannel(Options{Channel: *channel})
	case "season":
		cmd := flag.NewFlagSet("season", flag.ExitOnError)
		channel := cmd.String("channel", "", "Channel ID for stats")
		cmd.Parse(args[1:])
		if *channel == "" {
			cmd.Usage()
			os.Exit(1)
		}
		games, err := getGameList(*channel, "")
		if err != nil {
			log.Fatal(err)
		}
		start, end := seasonRangeForDate(time.Now())
		for _, game := range games {
			stats, err := GetStats(game, *channel, start.Format("2006-01-02"), end.Format("2006-01-02"))
			if len(stats) > 0 {
				fmt.Printf("### %s\n", game)
				if err != nil {
					log.Fatal(err)
				}
				PrintStats(stats, "md-discord")
			}

		}
	case "serve":
		cmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := cmd.String("port", "7654", "Port to run server")
		cmd.Parse(args[1:])
		addr := fmt.Sprintf(":%s", *port)
		logPrintln("Starting server on %s", addr)
		err = StartServer(addr)
	case "stats":
		cmd := flag.NewFlagSet("stats", flag.ExitOnError)
		game := cmd.String("game", "Wordle", "Game to print stats")
		channel := cmd.String("channel", "", "Channel ID for stats")
		format := cmd.String("format", "", "Format for output")
		cmd.Parse(args[1:])
		if *channel == "" || *game == "" {
			cmd.Usage()
			os.Exit(1)
		}
		stats, err := GetStats(*game, *channel, "", "")
		if err != nil {
			log.Fatal(err)
		}
		PrintStats(stats, *format)
	default:
		help()
	}
	if err != nil {
		log.Fatal(err)
	}
}
