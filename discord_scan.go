package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
func ScanChannel(options Options) error {
	// Create a new request
	params := url.Values{}
	if options.Before != "" {
		params.Add("before", options.Before)
		logPrintln("Scan channel <%s> before %s", options.Channel, options.Before)
	} else if options.After != "" {
		params.Add("after", options.After)
		logPrintln("Scan channel <%s> after %s", options.Channel, options.After)
	} else {
		logPrintln("Full rescan of channel <%s>", options.Channel)
	}
	baseURL := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages", options.Channel)
	url := baseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	authorization, err := getAuthorization()
	if err != nil {
		return err
	}
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
		err = fmt.Errorf("request failed with status: %s", resp.Status)
		return err
	}
	var messages []Message
	err = json.Unmarshal(body, &messages)
	if err != nil {
		return err
	}
	count := len(messages)
	if count == 0 {
		logPrintln("No new records found.")
		return nil
	}
	scores, err := ParseScores(messages)
	if err != nil {
		return err
	}
	err = AddScores(scores)
	if err != nil {
		return err
	}
	logPrintln("%d records updated (%s - %s)", count, messages[0].ID, messages[len(messages)-1].ID)
	// Check other pages
	// Unsure if assumption about message order is safe
	if options.Before != "" || options.Before == "" && options.After == "" {
		first_id := messages[count-1].ID
		prev_page := options
		prev_page.Before = first_id
		ScanChannel(prev_page)
	}
	if options.After != "" || options.Before == "" && options.After == "" {
		last_id := messages[0].ID
		next_page := options
		next_page.After = last_id
		ScanChannel(next_page)
	}
	return nil
}
