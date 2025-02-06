package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func LoadDatabase() (*sql.DB, error) {
	// Open database connection
	db, err := sql.Open("sqlite3", "./scores.db")
	if err != nil {
		return nil, err
	}
	// CREATE TABLE scores
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY,
			username TEXT,
			game TEXT,
			score TEXT,
            content TEXT
		)
	`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Add scores to database
func AddScores(db *sql.DB, scores []Score) error {
	// Prepare the upsert statement
	stmt, err := db.Prepare(`
		INSERT INTO scores (id, username, game, score, content)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			game = excluded.game,
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
		_, err := tx.Stmt(stmt).Exec(score.ID, score.Username, score.Game, score.Score, score.Content)
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