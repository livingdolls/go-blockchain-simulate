package models

type Candle struct {
	ID           int64   `db:"id" json:"id"`
	IntervalType string  `db:"interval_type" json:"interval_type"`
	StartTime    int64   `db:"start_time" json:"start_time"`
	OpenPrice    float64 `db:"open_price" json:"open_price"`
	HighPrice    float64 `db:"high_price" json:"high_price"`
	LowPrice     float64 `db:"low_price" json:"low_price"`
	ClosePrice   float64 `db:"close_price" json:"close_price"`
	Volume       float64 `db:"volume" json:"volume"`
}
