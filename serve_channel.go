package main

import (
	"html/template"
	"net/http"
)

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
	stats, err := GetStats(game, channelID, dateStart, dateEnd)
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
