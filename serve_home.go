package main

import "net/http"

// Handler for /
func rootHandler(w http.ResponseWriter, r *http.Request) {
	scores, err := GetRecentScores()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "home.tmpl", struct {
		Scores []Score
	}{
		Scores: scores,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
