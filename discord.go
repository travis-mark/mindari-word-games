package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	Before  int    // Backward search pointer
	After   int    // Forward search pointer
}

// TODO: Handle more than 50 messages when message count is high enough
func fetchFromDiscord(options Options) error {
	// Create a new request
	params := url.Values{}
	params.Add("limit", "50")
	if options.Before > 0 {
		params.Add("before", strconv.Itoa(options.Before))
	}
	if options.After > 0 {
		params.Add("after", strconv.Itoa(options.After))
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

	// Upsert to DB
	db, err := LoadDatabase()
	scores, err := ParseScores(messages)
	if err != nil {
		return err
	}
	err = AddScores(db, scores)
	if err != nil {
		return err
	}
	return nil
}
