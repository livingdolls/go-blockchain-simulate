package utils

import "time"

func FloorTime(timestamp int64, interval string) int64 {
	t := time.Unix(timestamp, 0)
	var floored time.Time

	switch interval {
	case "1m":
		floored = t.Truncate(time.Minute)
	case "5m":
		min := t.Truncate(time.Minute)
		floored = time.Date(min.Year(), min.Month(), min.Day(), min.Hour(), (min.Minute()/5)*5, 0, 0, min.Location())
	case "15m":
		min := t.Truncate(time.Minute)
		floored = time.Date(min.Year(), min.Month(), min.Day(), min.Hour(), (min.Minute()/15)*15, 0, 0, min.Location())
	case "30m":
		min := t.Truncate(time.Minute)
		floored = time.Date(min.Year(), min.Month(), min.Day(), min.Hour(), (min.Minute()/30)*30, 0, 0, min.Location())
	case "1h":
		floored = t.Truncate(time.Hour)
	case "4h":
		hour := t.Truncate(time.Hour)
		floored = time.Date(hour.Year(), hour.Month(), hour.Day(), (hour.Hour()/4)*4, 0, 0, 0, hour.Location())
	case "1d":
		floored = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	default:
		floored = t.Truncate(time.Minute)
	}

	return floored.Unix()
}

func IntervalDuration(interval string) int64 {
	switch interval {
	case "1m":
		return int64(time.Minute.Seconds())
	case "5m":
		return int64((5 * time.Minute).Seconds())
	case "15m":
		return int64((15 * time.Minute).Seconds())
	case "30m":
		return int64((30 * time.Minute).Seconds())
	case "1h":
		return int64(time.Hour.Seconds())
	case "4h":
		return int64((4 * time.Hour).Seconds())
	case "1d":
		return int64((24 * time.Hour).Seconds())
	default:
		return int64(time.Minute.Seconds())
	}
}
