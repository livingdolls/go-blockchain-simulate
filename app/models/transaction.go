package models

type Transaction struct {
	ID          int64   `db:"id"`
	FromAddress string  `db:"from_address"`
	ToAddress   string  `db:"to_address"`
	Amount      float64 `db:"amount"`
	Signature   string  `db:"signature"`
	Status      string  `db:"status"`
}
