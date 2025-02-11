package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type Stats struct {
	Username string
	Lowest   float32
	Average  float32
	Highest  float32
}

func GetStats(db *sql.DB, game string) ([]Stats, error) {
	sql := `
		SELECT username, MIN(score), AVG(score), MAX(score)
		FROM scores
		WHERE game = ?
		GROUP BY username
		ORDER BY 3
	`
	rows, err := db.Query(sql, game)
	if err != nil {
		return nil, fmt.Errorf("Failed to get stats: %v", err)
	}
	var stats []Stats
	for rows.Next() {
		var stat Stats
		err := rows.Scan(
			&stat.Username,
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
	fmt.Printf("Username\tLowest\tAverage\tHighest\n")
	for _, stat := range stats {
		fmt.Printf("%s\t%f\t%f\t%f\n", stat.Username, stat.Lowest, stat.Average, stat.Highest)
	}
}
