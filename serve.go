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

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Only redirect if the path is exactly "/"
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/stats", http.StatusFound)
		return
	}
	// Handle 404 for other non-existing paths
	http.NotFound(w, r)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	games, err := GetGames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := r.URL.Query().Get("game")
	if game == "" {
		game = games[0]
	}
	stats, err := GetStats(game)
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
	http.HandleFunc("/stats/", statsHandler)
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(addr, nil)
}
