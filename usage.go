package main

import (
	"os"
	"text/template"
)

type HelpParameters struct {
	ProgramName string
}

func help() {
	data := HelpParameters{ProgramName: os.Args[0]}
	const text = `
Mindari's Word Games is a tool to extract Wordle, etc... scores from a shared Discord channel.

Usage:

        {{ .ProgramName }} <command> [arguments]

The commands are:

        monitor     Periodically monitor a channel for posted scores
        rescan      Do a full rescan of a channel (in case of defects or edits)
        serve       Start a local webserver to show stats and a leaderboard
        stats       Print stats to standard output to use for custom graphs
		
`
	tmpl := template.Must(template.New("help").Parse(text))
	tmpl.Execute(os.Stdout, data)
}
