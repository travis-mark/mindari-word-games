package main

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.tmpl
var templateFS embed.FS
var tmpl = template.Must(template.ParseFS(templateFS, "*.tmpl"))

type StatsPageViewModel struct {
	CurrentGame string
	Games       []string
	Stats       []Stats
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	games, err := getGameList("", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := r.URL.Query().Get("game")
	if game == "" {
		game = games[0]
	}
	stats, err := GetStats(game, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vm := StatsPageViewModel{
		CurrentGame: game,
		Games:       games,
		Stats:       stats,
	}
	err = tmpl.ExecuteTemplate(w, "stats.tmpl", vm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func StartServer(addr string) error {
	http.HandleFunc("/c/", channelHandler)
	http.HandleFunc("/u/", userHandler)
	http.HandleFunc("/stats/", statsHandler)
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(addr, nil)
}
