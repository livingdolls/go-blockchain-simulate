package models

type Transaction struct {
	ID          int64   `db:"id"`
	FromAddress string  `db:"from_address"`
	ToAddress   string  `db:"to_address"`
	Amount      float64 `db:"amount"`
	Fee         float64 `db:"fee"`
	Signature   string  `db:"signature"`
	Status      string  `db:"status"`
}
