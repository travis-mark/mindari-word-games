package main

import (
	"html/template"
	"net/http"
	"strings"
)

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
	sql := `
		SELECT id, channel_id, username, s.game, s.game_number, score, win, hardmode
		FROM scores s
		JOIN puzzles p
			ON s.game = p.game AND s.game_number = p.game_number
		WHERE s.game = ? AND s.username = ? AND p.date >= ? AND p.date <= ?
		ORDER BY s.game_number DESC
	`
	rows, err := db.Query(sql, game, username, from, to)
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

func getFriendNames(username string) ([]string, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT DISTINCT username
		FROM scores
		WHERE channel_id IN (
			SELECT DISTINCT channel_id
			FROM scores
			WHERE username = ?)
	`
	rows, err := db.Query(sql, username)
	if err != nil {
		return nil, err
	}
	var friends []string
	for rows.Next() {
		var friend string
		err := rows.Scan(&friend)
		if err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return friends, nil
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
		BarMax:      getBarMaxValue(game),
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
