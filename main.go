package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

// Continue run until ^C. Used for server like discord commands that return but need to keep going.
func keepAlive() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logPrintln("Press Ctrl+C to exit")
	<-stop
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0][0] == '-' {
		args = append([]string{"help"}, args...)
	}
	cmd := args[0]
	var err error
	switch cmd {
	case "echo":
		dc, err := initDiscordConnection()
		if err != nil {
			log.Fatal(err)
		}
		_, err = dc.enableEchoCommand()
		if err != nil {
			log.Fatal(err)
		}
		keepAlive()
		dc.close()
	case "monitor":
		dc, err := initDiscordConnection()
		if err != nil {
			log.Fatal(err)
		}
		err = dc.startDiscordMonitor()
		if err != nil {
			log.Fatal(err)
		}
		keepAlive()
	case "rescan":
		cmd := flag.NewFlagSet("rescan", flag.ExitOnError)
		channel := cmd.String("channel", "", "Channel ID to scan")
		cmd.Parse(args[1:])
		if *channel == "" {
			cmd.Usage()
			os.Exit(1)
		}
		dc, err := initDiscordConnection()
		if err != nil {
			log.Fatal(err)
		}
		err = dc.scanChannel(Options{Channel: *channel})
		if err != nil {
			log.Fatal(err)
		}
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
