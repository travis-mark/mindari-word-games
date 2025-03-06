package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Stats struct {
	Username string
	Count    int
	Lowest   float32
	Average  float32
	Highest  float32
}

// Get a list of games. Add channelID or username to filter the list.
func getGameList(channelID string, username string) ([]string, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	var rows *sql.Rows
	if channelID != "" && username != "" {
		sql := `SELECT DISTINCT game FROM scores WHERE channel_id = ? AND username = ?`
		rows, err = db.Query(sql, channelID, username)
	} else if channelID != "" {
		sql := `SELECT DISTINCT game FROM scores WHERE channel_id = ?`
		rows, err = db.Query(sql, channelID)
	} else if username != "" {
		sql := `SELECT DISTINCT game FROM scores WHERE username = ?`
		rows, err = db.Query(sql, username)
	} else {
		sql := `SELECT DISTINCT game FROM scores`
		rows, err = db.Query(sql)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %v", err)
	}
	var games []string
	for rows.Next() {
		var game string
		err := rows.Scan(&game)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return games, nil
}

func GetStats(game string, channelID string, from string, to string) ([]Stats, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	if from == "" {
		from = defaultDateStart()
	}
	if to == "" {
		to = defaultDateEnd()
	}
	var rows *sql.Rows
	sql := `
		SELECT username, COUNT(id), MIN(score), AVG(score), MAX(score)
		FROM scores s
		JOIN puzzles p
			ON s.game = p.game AND s.game_number = p.game_number
		WHERE s.game = ? AND channel_id = ? AND p.date >= ? AND p.date <= ?
		GROUP BY username
		ORDER BY 4
	`
	rows, err = db.Query(sql, game, channelID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %v", err)
	}
	var stats []Stats
	for rows.Next() {
		var stat Stats
		err := rows.Scan(
			&stat.Username,
			&stat.Count,
			&stat.Lowest,
			&stat.Average,
			&stat.Highest,
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

func PrintStats(stats []Stats, format string) {
	fmt.Print(SPrintStats(stats, format))
}

func SPrintStatsMarkdownDiscord(stats []Stats) string {
	usernameColumnTitle := "Username"
	usernameColumnSize := len(usernameColumnTitle)
	for _, stat := range stats {
		if len(stat.Username) > usernameColumnSize {
			usernameColumnSize = len(stat.Username)
		}
	}
	usernameColumnTitle = fmt.Sprintf("%-*s", usernameColumnSize, usernameColumnTitle)
	var builder strings.Builder
	builder.WriteString("```md\n")
	header := fmt.Sprintf("| %s |  # | Min | Mean | Max |\n", usernameColumnTitle)
	builder.WriteString(header)
	linebreak := fmt.Sprintf("| %s | -- | --- | ---- | --- |\n", strings.Repeat("-", usernameColumnSize))
	builder.WriteString(linebreak)
	for _, stat := range stats {
		s := fmt.Sprintf("| %-*s | %2d | %3.0f | %4.1f | %3.0f |\n", usernameColumnSize, stat.Username, stat.Count, stat.Lowest, stat.Average, stat.Highest)
		builder.WriteString(s)
	}
	builder.WriteString("```\n")
	return builder.String()
}

func SPrintStatsTabs(stats []Stats) string {
	var builder strings.Builder
	builder.WriteString("Username\tGames\tLowest\tAverage\tHighest\n")
	for _, stat := range stats {
		s := fmt.Sprintf("%s\t%d\t%0.0f\t%0.2f\t%0.0f\n", stat.Username, stat.Count, stat.Lowest, stat.Average, stat.Highest)
		builder.WriteString(s)
	}
	return builder.String()
}

func SPrintStats(stats []Stats, format string) string {
	switch format {
	case "md-discord":
		return SPrintStatsMarkdownDiscord(stats)
	default:
		return SPrintStatsTabs(stats)
	}
}
