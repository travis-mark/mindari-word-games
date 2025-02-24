package main

import (
	"os"
	"text/template"
)

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
		ExecName: AppExecName(),
		FullName: AppFullName(),
	})
}
