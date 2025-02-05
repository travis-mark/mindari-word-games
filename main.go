package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// Message represents a Discord message
type Message struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
	Author  Author `json:"author"`
	// Add other fields as needed
}

// Author represents the message author
type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Score struct {
	ID       string
	Username string
	Score    string
	Content  string
}

// Add scores to database
// TODO: Normalize author
func AddScores(db *sql.DB, scores []Score) error {
	// Prepare the upsert statement
	stmt, err := db.Prepare(`
		INSERT INTO scores (id, username, score, content)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			score = excluded.score,
            content = excluded.content
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Execute the upsert for each score
	for _, score := range scores {
		_, err := tx.Stmt(stmt).Exec(score.ID, score.Username, score.Score, score.Content)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed add score %s: %v", score.ID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Parse message to extract score
// TODO: Parse game
// TODO: Parse score
func ParseScores(messages []Message) ([]Score, error) {
	scores := make([]Score, 0, len(messages))

	for _, msg := range messages {
		if msg.Type != 0 {
			continue
		}
		lines := strings.Split(msg.Content, "\n")
		if len(lines) == 0 {
			continue
		}
		score := Score{
			ID:       msg.ID,
			Username: msg.Author.Username,
			Score:    lines[0],
			Content:  msg.Content,
		}
		scores = append(scores, score)
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("no valid scores found in messages")
	}

	return scores, nil
}

func main() {
	// Load .env to get secrets
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new request
	// TODO: Handle more than 50 messages
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

	// Open database connection
	db, err := sql.Open("sqlite3", "./scores.db")
	if err != nil {
		log.Fatal(err)
	}
	// Create the users table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY,
			username TEXT,
			score TEXT,
            content TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
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
