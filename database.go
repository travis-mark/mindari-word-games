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
			game_number TEXT,
			score TEXT,
            content TEXT,
			win TEXT,
			hardmode TEXT
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
		INSERT INTO scores (id, username, game, game_number, score, content, win, hardmode)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			game = excluded.game,
			game_number = excluded.game_number,
			score = excluded.score,
            content = excluded.content,
			win = excluded.win,
			hardmode = excluded.hardmode
	`)
	if err != nil {
		return fmt.Errorf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v", err)
	}

	// Execute the upsert for each score
	for _, score := range scores {
		_, err := tx.Stmt(stmt).Exec(score.ID, score.Username, score.Game, score.GameNumber, score.Score, score.Content, score.Win, score.Hardmode)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Failed to add score %s: %v", score.ID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}

	return nil
}
