package models

type BalanceDiscrepancy struct {
	ID              int64   `db:"id"`
	Address         string  `db:"address"`
	BlockNumber     int     `db:"block_number"`
	ExpectedBalance float64 `db:"expected_balance"`
	ActualBalance   float64 `db:"actual_balance"`
	Difference      float64 `db:"difference"`
	Resolved        bool    `db:"resolved"`
	ResolutionNote  string  `db:"resolution_note"`
	Timestamp       int64   `db:"timestamp"`
}
