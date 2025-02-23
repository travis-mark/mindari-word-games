package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Stats struct {
	Username string
	Count    int
	Lowest   float32
	Average  float32
	Highest  float32
}

func GetGames() ([]string, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	sql := `
		SELECT DISTINCT game
		FROM scores
	`
	rows, err := db.Query(sql)
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

func GetStats(game string, channelID string) ([]Stats, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	var rows *sql.Rows
	if channelID != "" {
		sql := `
			SELECT username, COUNT(id), MIN(score), AVG(score), MAX(score)
			FROM scores
			WHERE game = ? and channel_id = ?
			GROUP BY username
			ORDER BY 4
		`
		rows, err = db.Query(sql, game, channelID)
	} else {
		sql := `
			SELECT username, COUNT(id), MIN(score), AVG(score), MAX(score)
			FROM scores
			WHERE game = ?
			GROUP BY username
			ORDER BY 4
		`
		rows, err = db.Query(sql, game)
	}
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

func PrintStats(stats []Stats) {
	fmt.Printf("Username\tGames\tLowest\tAverage\tHighest\n")
	for _, stat := range stats {
		fmt.Printf("%s\t%d\t%0.0f\t%0.2f\t%0.0f\n", stat.Username, stat.Count, stat.Lowest, stat.Average, stat.Highest)
	}
}
