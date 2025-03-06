package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func initDatabase() (*sql.DB, error) {
	// Open database connection
	db, err := sql.Open("sqlite3", "./scores.db")
	if err != nil {
		return nil, err
	}
	// Scores
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id TEXT UNIQUE,
			channel_id TEXT,
			username TEXT,
			game TEXT,
			game_number TEXT,
			score TEXT,
			win TEXT,
			hardmode TEXT,
			UNIQUE (username, game, game_number, score)
		)
	`)
	if err != nil {
		return nil, err
	}
	// Channels
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			channel_id TEXT UNIQUE,
			guild_id TEXT,
			name TEXT
		)
	`)
	if err != nil {
		return nil, err
	}
	// Puzzles
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS puzzles (
			game TEXT,
			game_number TEXT,
			date TEXT,
			solution TEXT,
			UNIQUE (game, game_number)
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
	db, err := initDatabase()
	_db = db
	return db, err
}

// Add scores to database
func AddScores(scores []Score) error {
	db, err := GetDatabase()
	if err != nil {
		return err
	}
	score_stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO scores (id, channel_id, username, game, game_number, score, win, hardmode)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare score statement: %v", err)
	}
	defer score_stmt.Close()
	puzzle_stmt, err := db.Prepare(`
		INSERT OR IGNORE INTO puzzles (game, game_number, date)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare puzzle statement: %v", err)
	}
	defer score_stmt.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	// Execute the upsert for each score
	for _, score := range scores {
		_, err := tx.Stmt(score_stmt).Exec(score.ID, score.ChannelID, score.Username, score.Game, score.GameNumber, score.Score, score.Win, score.Hardmode)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add score %s: %v", score.ID, err)
		}
		date, err := dateFromDiscordSnowflake(score.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get snowflake %s: %v", score.ID, err)
		}
		_, err = tx.Stmt(puzzle_stmt).Exec(score.Game, score.GameNumber, date)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add puzzle %s: %v", score.ID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
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

// Get latest scores
func GetRecentScores() ([]Score, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT id, channel_id, username, game, game_number, score, win, hardmode
		FROM scores
		ORDER BY id DESC
		LIMIT 5
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent scores: %v", err)
	}
	var scores []Score
	for rows.Next() {
		var score Score
		err := rows.Scan(
			&score.ID,
			&score.ChannelID,
			&score.Username,
			&score.Game,
			&score.GameNumber,
			&score.Score,
			&score.Win,
			&score.Hardmode,
		)
		if err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return scores, nil
}
