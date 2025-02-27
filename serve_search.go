package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type SearchResult struct {
	Type string
	ID string
	Name string
}

func FindByChannelOrUsername(query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT "channel", channel_id, name 
		FROM channels
		WHERE name LIKE ?
		LIMIT 50
	`
	rows, err := db.Query(sql, "%" + query + "%")
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %v", err)
	}
	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		err := rows.Scan(
			&result.Type,
			&result.ID,
			&result.Name,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		if len(results) >= 50 {
			return results, nil
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	sql = `
		SELECT "user", username, username 
		FROM (SELECT DISTINCT username FROM scores)
		WHERE username LIKE ?
		LIMIT 50
	`
	rows, err = db.Query(sql, "%" + query + "%")
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	for rows.Next() {
		var result SearchResult
		err := rows.Scan(
			&result.Type,
			&result.ID,
			&result.Name,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		if len(results) >= 50 {
			return results, nil
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// Handler for /?q=
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	results, err := FindByChannelOrUsername(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "search.tmpl", struct {
		Query		string
		Results		[]SearchResult
		Style       template.CSS
	}{
		Query: query,
		Results: results,
		Style:       template.CSS(stylesheet),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
