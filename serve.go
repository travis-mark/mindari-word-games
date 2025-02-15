package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var templates = template.Must(template.ParseFiles("scan.tpml", "stats.tpml"))

type WordGameServer struct {
	db *sql.DB
}

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

func (svr *WordGameServer) scanHandler(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 3 {
		http.Error(w, fmt.Errorf("Channel ID missing from path /scan/<channel_id>").Error(), http.StatusBadRequest)
	}
	var buffer bytes.Buffer
	out := log.New(&buffer, "", log.Ltime)
	channel := pathSegments[2]
	if channel != "" {
		FetchFromDiscordAndPersist(svr.db, out, Options{Channel: channel})
	}
	err := templates.ExecuteTemplate(w, "scan.tpml", buffer.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	http.HandleFunc("/scan/", svr.scanHandler)
	http.HandleFunc("/stats/", svr.statsHandler)
	http.HandleFunc("/", rootHandler)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
