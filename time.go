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
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return thirtyDaysAgo.Format("2006-01-02")
}

func defaultDateEnd() string {
	today := time.Now()
	return today.Format("2006-01-02")
}

func seasonRangeForDate(t time.Time) (time.Time, time.Time) {
	t0 := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	t1 := t0.AddDate(0, 1, 0).AddDate(0, 0, -1)
	return t0, t1
}
