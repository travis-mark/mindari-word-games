package main

import (
	"database/sql"
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseFiles("stats.tpml"))

type WordGameServer struct {
	db *sql.DB
}

type StatsPageViewModel struct {
	CurrentGame string
	Games       []string
	Stats       []Stats
}

func (svr *WordGameServer) statsHandler(w http.ResponseWriter, r *http.Request) {
	games, err := GetGames(svr.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := r.URL.Query().Get("game")
	if game == "" {
		game = games[0]
	}
	stats, err := GetStats(svr.db, game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vm := StatsPageViewModel{
		CurrentGame: game,
		Games:       games,
		Stats:       stats,
	}
	err = templates.ExecuteTemplate(w, "stats.tpml", vm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func startServer(db *sql.DB, addr string) error {
	svr := WordGameServer{db: db}
	http.HandleFunc("/stats", svr.statsHandler)
	return http.ListenAndServe(addr, nil)
}
