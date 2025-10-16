package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Open database connection
func initDatabase() (*sql.DB, error) {
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
func getDatabase() (*sql.DB, error) {
	if _db != nil {
		return _db, nil
	}
	db, err := initDatabase()
	_db = db
	return db, err
}

// Add scores to database
func addScores(scores []Score) error {
	db, err := getDatabase()
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

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
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
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Grab oldest and newest id. Used to download incrementally.
func getScoreIDRange() (string, string, error) {
	db, err := getDatabase()
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

// Get most recent message ID for a specific channel
func getMostRecentMessageID(channelID string) (string, error) {
	db, err := getDatabase()
	if err != nil {
		return "", err
	}
	var messageID string
	err = db.QueryRow("SELECT MAX(id) FROM scores WHERE channel_id = ?", channelID).Scan(&messageID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return messageID, nil
}

// Get latest scores
func GetRecentScores() ([]Score, error) {
	db, err := getDatabase()
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

type SearchResult struct {
	Type string
	ID   string
	Name string
}

// Used by search box
func findByChannelOrUsername(query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT "channel", channel_id, name 
		FROM channels
		WHERE name LIKE ?
		LIMIT 50
	`
	rows, err := db.Query(sql, "%"+query+"%")
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
	rows, err = db.Query(sql, "%"+query+"%")
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

func getScoresByUser(game string, username string, from string, to string) ([]Score, error) {
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT id, channel_id, username, s.game, s.game_number, score, win, hardmode
		FROM scores s
		JOIN puzzles p
			ON s.game = p.game AND s.game_number = p.game_number
		WHERE s.game = ? AND s.username = ? AND p.date >= ? AND p.date <= ?
		ORDER BY s.game_number DESC
	`
	rows, err := db.Query(sql, game, username, from, to)
	if err != nil {
		return nil, err
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

// List usernames in same guild as user. Used for user dropdown.
func getFriendNames(username string) ([]string, error) {
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT DISTINCT username
		FROM scores
		WHERE channel_id IN (
			SELECT DISTINCT channel_id
			FROM scores
			WHERE username = ?)
	`
	rows, err := db.Query(sql, username)
	if err != nil {
		return nil, err
	}
	var friends []string
	for rows.Next() {
		var friend string
		err := rows.Scan(&friend)
		if err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return friends, nil
}

type AttendanceStats struct {
	Username    string
	GamesPlayed int
	DaysActive  int
}

// Get attendance statistics for a guild for a specific month - games played and days active per player
func getAttendanceStatsForMonth(guildID string, month string) ([]AttendanceStats, error) {
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT 
			s.username,
			COUNT(s.id) as games_played,
			COUNT(DISTINCT DATE(p.date)) as days_active
		FROM scores s
		JOIN puzzles p ON s.game = p.game AND s.game_number = p.game_number
		JOIN channels c ON s.channel_id = c.channel_id
		WHERE c.guild_id = ? AND strftime('%Y-%m', p.date) = ?
		GROUP BY s.username
		ORDER BY s.username
	`
	rows, err := db.Query(sql, guildID, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance stats: %v", err)
	}
	defer rows.Close()

	var stats []AttendanceStats
	for rows.Next() {
		var stat AttendanceStats
		err := rows.Scan(
			&stat.Username,
			&stat.GamesPlayed,
			&stat.DaysActive,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
