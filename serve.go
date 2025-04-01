package main

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
)

//go:embed *.tmpl
var templateFS embed.FS
var tmpl = template.Must(template.ParseFS(templateFS, "*.tmpl"))

//go:embed style.css
var stylesheet string

// Handler for /
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("q") != "" {
		searchHandler(w, r)
		return
	}
	scores, err := GetRecentScores()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "home.tmpl", struct {
		AppFullName string
		Scores      []Score
		Style       template.CSS
	}{
		AppFullName: appFullName(),
		Scores:      scores,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /?q=
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	results, err := findByChannelOrUsername(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "search.tmpl", struct {
		Query   string
		Results []SearchResult
		Style   template.CSS
	}{
		Query:   query,
		Results: results,
		Style:   template.CSS(stylesheet),
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /channel
func channelHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	channelID := params.Get("id")
	if channelID == "" {
		http.Error(w, "Channel ID Required", http.StatusInternalServerError)
		return
	}
	channel, err := readChannelInfo(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	from := params.Get("from")
	if from == "" {
		from = defaultDateStart()
	}
	to := params.Get("to")
	if to == "" {
		to = defaultDateEnd()
	}
	games, err := getGameList(channel.GuildID, "", from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := params.Get("game")
	if game == "" {
		game = games[0]
	}
	stats, err := getStats(game, channel.GuildID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "channel.tmpl", struct {
		ChannelID   string
		ChannelName string
		CurrentGame string
		DateStart   string
		DateEnd     string
		Games       []string
		Stats       []Stats
		Style       template.CSS
	}{
		ChannelID:   channel.ID,
		ChannelName: channel.Name,
		CurrentGame: game,
		DateStart:   from,
		DateEnd:     to,
		Games:       games,
		Stats:       stats,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /stats
func statsHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	channelID := params.Get("cid")
	if channelID == "" {
		http.Error(w, "Channel Required", http.StatusInternalServerError)
		return
	}
	channel, err := readChannelInfo(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	from := params.Get("from")
	if from == "" {
		from = defaultDateStart()
	}
	to := params.Get("to")
	if to == "" {
		to = defaultDateEnd()
	}
	games, err := getGameList(channel.GuildID, "", from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type GameStats struct {
		Game  string
		Stats []Stats
	}
	var gameStats []GameStats
	for _, game := range games {
		stats, err := getStats(game, channel.GuildID, from, to)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(stats) > 0 {
			gameStats = append(gameStats, GameStats{Game: game, Stats: stats})
		}
	}
	err = tmpl.ExecuteTemplate(w, "stats.tmpl", struct {
		ChannelID   string
		ChannelName string
		From        string
		To          string
		GameStats   []GameStats
		Style       template.CSS
	}{
		ChannelID:   channelID,
		ChannelName: channel.Name,
		From:        from,
		To:          to,
		GameStats:   gameStats,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /user
func userHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	username := params.Get("name")
	if username == "" {
		http.Error(w, "Username Required", http.StatusInternalServerError)
		return
	}
	from := params.Get("from")
	if from == "" {
		from = defaultDateStart()
	}
	to := params.Get("to")
	if to == "" {
		to = defaultDateEnd()
	}
	games, err := getGameList("", username, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := params.Get("game")
	if game == "" {
		game = games[0]
	}
	friends, err := getFriendNames(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	scores, err := getScoresByUser(game, username, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var barMax int
	switch {
	case strings.Contains(game, "Octordle"):
		barMax = 100
	default:
		barMax = 7
	}
	err = tmpl.ExecuteTemplate(w, "user.tmpl", struct {
		Username    string
		BarMax      int
		CurrentGame string
		DateStart   string
		DateEnd     string
		Friends     []string
		Games       []string
		Scores      []Score
		Style       template.CSS
	}{
		Username:    username,
		BarMax:      barMax,
		CurrentGame: game,
		DateStart:   from,
		DateEnd:     to,
		Friends:     friends,
		Games:       games,
		Scores:      scores,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func startWebServer(addr string) error {
	http.HandleFunc("/channel", channelHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(addr, nil)
}
