package main

import (
	"net/http"
	"regexp"
)

var channelRegex = regexp.MustCompile(`^/c/([^/]+)(?:/.*)?$`)

// Handler for /c/<CHANNEL_ID>
func channelHandler(w http.ResponseWriter, r *http.Request) {
	matches := channelRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		// No match found
		http.Error(w, "Invalid channel URL", http.StatusBadRequest)
		return
	}
	channelID := matches[1]
	channel, err := ReadChannelInfo(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	games, err := getGameList(channelID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := r.URL.Query().Get("game")
	if game == "" {
		game = games[0]
	}
	stats, err := GetStats(game, channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "channel.tmpl", struct {
		ChannelID   string
		ChannelName string
		CurrentGame string
		Games       []string
		Stats       []Stats
	}{
		ChannelID:   channel.ID,
		ChannelName: channel.Name,
		CurrentGame: game,
		Games:       games,
		Stats:       stats,
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
