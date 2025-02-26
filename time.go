package main

import "time"

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

func defaultDateStart() string {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return thirtyDaysAgo.Format("2006-01-02")
}

func defaultDateEnd() string {
	today := time.Now()
	return today.Format("2006-01-02")
}