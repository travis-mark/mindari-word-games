package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Message represents a Discord message
type Message struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
	Author  Author `json:"author"`
}

// Author represents the message author
type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Options struct {
	Channel string // Discord Channel ID
	Before  string // Backward search pointer
	After   string // Forward search pointer
}

// Fetch messages from Discord, parse for puzzles and save to DB
//
// Ref: https://discord.com/developers/docs/resources/message#get-channel-messages
func FetchFromDiscordAndPersist(db *sql.DB, options Options) error {
	// Create a new request
	params := url.Values{}
	if options.Before != "" {
		params.Add("before", options.Before)
	}
	if options.After != "" {
		params.Add("after", options.After)
	}
	baseURL := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages", options.Channel)
	url := baseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	authorization := os.Getenv("AUTHORIZATION")
	req.Header.Set("Authorization", authorization)
	req.Header.Set("User-Agent", "Mindari Word Games (0.0-alpha)")

	// Create HTTP client and execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Request failed with status: %s", resp.Status)
		return err
	}
	var messages []Message
	err = json.Unmarshal(body, &messages)
	if err != nil {
		return err
	}
	count := len(messages)
	if count == 0 {
		return nil
	}

	// Upsert to DB
	scores, err := ParseScores(messages)
	if err != nil {
		return err
	}
	err = AddScores(db, scores)
	if err != nil {
		return err
	}
	fmt.Printf("%d records updated (%s - %s)\n", count, messages[0].ID, messages[len(messages)-1].ID)

	// Check other pages
	// Unsure if assumption about message order is safe
	if options.Before != "" || options.Before == "" && options.After == "" {
		first_id := messages[count-1].ID
		prev_page := options
		prev_page.Before = first_id
		FetchFromDiscordAndPersist(db, prev_page)
	}
	if options.After != "" || options.Before == "" && options.After == "" {
		last_id := messages[0].ID
		next_page := options
		next_page.After = last_id
		FetchFromDiscordAndPersist(db, next_page)
	}

	return nil
}
