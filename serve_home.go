package main

import (
	"html/template"
	"net/http"
)

// Handler for /
func rootHandler(w http.ResponseWriter, r *http.Request) {
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
		AppFullName: AppFullName(),
		Scores:      scores,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
