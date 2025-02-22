package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// TODO: Restructure / make this private
func LoadDatabase() (*sql.DB, error) {
	// Open database connection
	db, err := sql.Open("sqlite3", "./scores.db")
	if err != nil {
		return nil, err
	}
	// CREATE TABLE scores
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id TEXT UNIQUE,
			username TEXT,
			game TEXT,
			game_number TEXT,
			score TEXT,
            content TEXT,
			win TEXT,
			hardmode TEXT,
			UNIQUE (username, game, game_number, score)
		)
	`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

var _db *sql.DB

// Get a connection to database. Reuses a shared connection if one is available.
func GetDatabase() (*sql.DB, error) {
	if _db != nil {
		return _db, nil
	}
	db, err := LoadDatabase()
	_db = db
	return db, err
}

// Add scores to database
// TODO: Drop content column from DB
func AddScores(scores []Score) error {
	db, err := GetDatabase()
	if err != nil {
		return err
	}
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO scores (id, username, game, game_number, score, content, win, hardmode)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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

// Grab oldest and newest id. Used to download incrementally.
func GetScoreIDRange() (string, string, error) {
	db, err := GetDatabase()
	if err != nil {
		return "", "", err
	}
	var oldest string
	var newest string
	err = db.QueryRow("SELECT MIN(id), MAX(id) FROM scores").Scan(&oldest, &newest)
	if err != nil {
		return "", "", err
	} else {
		return oldest, newest, nil
	}
}
