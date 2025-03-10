package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
)

// Shown in usage
func appExecName() string { return filepath.Base(os.Args[0]) }

// Shown in usage and web page titles
func appFullName() string { return "Mindari's Word Games" }

// Used by logger
func appInitials() string { return "MWG" }

// Usage statement
func help() {
	const text = `{{ .FullName }} is a tool to extract Wordle, etc... scores from a shared Discord channel.

Usage:

        {{ .ExecName }} <command> [arguments]

The commands are:

        help		Show this list
        monitor     Periodically monitor a channel for posted scores
        rescan      Do a full rescan of a channel (in case of defects or edits)
        serve       Start a local webserver to show stats and a leaderboard
        stats       Print stats to standard output to use for custom graphs
		
`
	tmpl := template.Must(template.New("help").Parse(text))
	tmpl.Execute(os.Stdout, struct {
		ExecName string
		FullName string
	}{
		ExecName: appExecName(),
		FullName: appFullName(),
	})
}

// Logging wrapper - Includes file and line number
func logPrintln(format string, v ...any) {
	// Capitalization for variables in final print statement
	App := appInitials()
	Message := fmt.Sprintf(format, v...)
	Template := "[%s] %s:%d:%s %s\n"
	pc, file, Line, ok := runtime.Caller(1)
	// This shouldn't fail, but do not swallow message if it does
	if !ok {
		log.Printf(Template, App, "UNKNOWN", 0, "UNKNOWN", Message)
	} else {
		fnWithModule := runtime.FuncForPC(pc).Name()
		fnParts := strings.Split(fnWithModule, ".")
		// Filename without path
		Name := filepath.Base(file)
		// Function name without module
		Fn := fnParts[len(fnParts)-1]
		log.Printf(Template, App, Name, Line, Fn, Message)
	}
}

// Continue run until ^C. Used for server like discord commands that return but need to keep going.
func keepAlive() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logPrintln("Press Ctrl+C to exit")
	<-stop
}

// Main entry point
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
			stats, err := getStats(game, *channel, start.Format("2006-01-02"), end.Format("2006-01-02"))
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
		err = startWebServer(addr)
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
		stats, err := getStats(*game, *channel, "", "")
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
