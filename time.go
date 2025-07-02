package main

import (
	"strconv"
	"time"
)

func dateToDiscordSnowflake(dateStr string) (int64, error) {
	date, err := time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		return 0, err
	}
	discordEpoch := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	milliseconds := date.UnixMilli() - discordEpoch.UnixMilli()
	snowflake := milliseconds << 22
	return snowflake, nil
}

func dateFromDiscordSnowflake(snowflake string) (string, error) {
	snowflakeInt, err := strconv.ParseInt(snowflake, 10, 64)
	if err != nil {
		return "", err
	}
	milliseconds := snowflakeInt >> 22
	discordEpoch := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	date := time.UnixMilli(discordEpoch.UnixMilli() + milliseconds)
	dateStr := date.Format("2006-01-02")
	return dateStr, nil
}

func defaultDateStart() string {
	t := time.Now()
	t0 := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	return t0.Format("2006-01-02")
}

func defaultDateEnd() string {
	t := time.Now()
	t0 := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	t1 := t0.AddDate(0, 1, 0).AddDate(0, 0, -1)
	return t1.Format("2006-01-02")
}

func getCurrentMonth() string {
	return time.Now().Format("2006-01")
}

func parseMonth(monthStr string) (time.Time, error) {
	return time.Parse("2006-01", monthStr)
}

func getPreviousMonth(monthStr string) string {
	t, err := parseMonth(monthStr)
	if err != nil {
		return getCurrentMonth()
	}
	prev := t.AddDate(0, -1, 0)
	return prev.Format("2006-01")
}

func getNextMonth(monthStr string) string {
	t, err := parseMonth(monthStr)
	if err != nil {
		return getCurrentMonth()
	}
	next := t.AddDate(0, 1, 0)
	return next.Format("2006-01")
}

func formatMonthDisplay(monthStr string) string {
	t, err := parseMonth(monthStr)
	if err != nil {
		return monthStr
	}
	return t.Format("January 2006")
}
