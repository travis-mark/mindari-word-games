package main

import (
	"html/template"
	"net/http"
	"regexp"
)

var userRegex = regexp.MustCompile(`^/u/([^/]+)(?:/.*)?$`)

func getScoresByUser(game string, username string) ([]Score, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT id, channel_id, username, game, game_number, score, win, hardmode
		FROM scores
		WHERE game = ? and username = ?
		ORDER BY game_number DESC
	`
	rows, err := db.Query(sql, game, username)
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
	scores, err := getScoresByUser(game, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "user.tmpl", struct {
		Username    string
		CurrentGame string
		Games       []string
		Scores      []Score
		Style       template.CSS
	}{
		Username:    username,
		CurrentGame: game,
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
