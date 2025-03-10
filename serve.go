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
		appFullName string
		Scores      []Score
		Style       template.CSS
	}{
		appFullName: appFullName(),
		Scores:      scores,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
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
	games, err := getGameList(channelID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := params.Get("game")
	if game == "" {
		game = games[0]
	}
	dateStart := params.Get("from")
	if dateStart == "" {
		dateStart = defaultDateStart()
	}
	dateEnd := params.Get("to")
	if dateEnd == "" {
		dateEnd = defaultDateEnd()
	}
	stats, err := getStats(game, channelID, dateStart, dateEnd)
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
		DateStart:   dateStart,
		DateEnd:     dateEnd,
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

// Handler for /user
func userHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	username := params.Get("name")
	if username == "" {
		http.Error(w, "Username Required", http.StatusInternalServerError)
		return
	}
	games, err := getGameList("", username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := params.Get("game")
	if game == "" {
		game = games[0]
	}
	from := params.Get("from")
	if from == "" {
		from = defaultDateStart()
	}
	to := params.Get("to")
	if to == "" {
		to = defaultDateEnd()
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
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(addr, nil)
}
