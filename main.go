package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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

func main() {
	// Load .env to get secrets
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new request
	// TODO: Handle more than 50 messages
	// https://discord.com/developers/docs/resources/message#get-channel-messages
	// When message count is high enough
	channel := os.Getenv("CHANNEL")
	if channel == "" {
		log.Fatal(fmt.Errorf("CHANNEL not set in environment"))
	}
	url := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages?limit=50", channel)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set headers
	authorization := os.Getenv("AUTHORIZATION")
	req.Header.Set("Authorization", authorization)
	req.Header.Set("User-Agent", "Mindari Word Games (0.0-alpha)")

	// Create HTTP client and execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status: %s\n", resp.Status)
	}

	// Parse JSON
	var messages []Message
	err = json.Unmarshal(body, &messages)
	if err != nil {
		log.Fatal(fmt.Sprintf("JSON parse error: %v", err))
	}

	db, err := LoadDatabase()

	// Upsert to DB
	scores, err := ParseScores(messages)
	if err != nil {
		log.Fatal(err)
	}
	err = AddScores(db, scores)
	if err != nil {
		log.Fatal(err)
	}
}
