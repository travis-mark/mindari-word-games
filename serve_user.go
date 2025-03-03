package main

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

var userRegex = regexp.MustCompile(`^/u/([^/]+)(?:/.*)?$`)

func getBarMaxValue(game string) int {
	switch {
	case strings.Contains(game, "Octordle"):
		return 100
	default:
		return 7
	}
}

func getScoresByUser(game string, username string, from string, to string) ([]Score, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	d_from, err := dateToDiscordSnowflake(from + "T00:00:00")
	if err != nil {
		return nil, err
	}
	d_to, err := dateToDiscordSnowflake(to + "T23:59:59")
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT id, channel_id, username, game, game_number, score, win, hardmode
		FROM scores
		WHERE game = ? AND username = ? AND id >= ? AND id <= ?
		ORDER BY game_number DESC
	`
	rows, err := db.Query(sql, game, username, d_from, d_to)
	if err != nil {
		return nil, err
	}
	var scores []Score
	for rows.Next() {
		var score Score
		err := rows.Scan(
			&score.ID,
			&score.ChannelID,
			&score.Username,
			&score.Game,
			&score.GameNumber,
			&score.Score,
			&score.Win,
			&score.Hardmode,
		)
		if err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return scores, nil

}

// Handler for /u/<USERNAME>
func userHandler(w http.ResponseWriter, r *http.Request) {
	matches := userRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		// No match found
		http.Error(w, "Invalid user URL", http.StatusBadRequest)
		return
	}
	username := matches[1]
	games, err := getGameList("", username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game := r.URL.Query().Get("game")
	if game == "" {
		game = games[0]
	}
	dateStart := r.URL.Query().Get("from")
	if dateStart == "" {
		dateStart = defaultDateStart()
	}
	dateEnd := r.URL.Query().Get("to")
	if dateEnd == "" {
		dateEnd = defaultDateEnd()
	}
	scores, err := getScoresByUser(game, username, dateStart, dateEnd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "user.tmpl", struct {
		Username    string
		CurrentGame string
		DateStart   string
		DateEnd     string
		Games       []string
		Scores      []Score
		Style       template.CSS
		BarMax      int
	}{
		Username:    username,
		CurrentGame: game,
		DateStart:   dateStart,
		DateEnd:     dateEnd,
		Games:       games,
		Scores:      scores,
		Style:       template.CSS(stylesheet),
		BarMax:      getBarMaxValue(game),
	})
	if err != nil {
		logPrintln("%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
