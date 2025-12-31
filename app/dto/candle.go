package dto

type CandleDTO struct {
	ID           int64   `json:"id"`
	IntervalType string  `json:"interval_type"`
	StartTime    int64   `json:"start_time"`
	OpenPrice    float64 `json:"open_price"`
	HighPrice    float64 `json:"high_price"`
	LowPrice     float64 `json:"low_price"`
	ClosePrice   float64 `json:"close_price"`
	Volume       float64 `json:"volume"`
}

func IsValidInterval(intervalType string) bool {
	validIntervals := map[string]bool{
		"1m":  true,
		"5m":  true,
		"15m": true,
		"30m": true,
		"1h":  true,
		"4h":  true,
		"1d":  true,
	}

	return validIntervals[intervalType]
}
