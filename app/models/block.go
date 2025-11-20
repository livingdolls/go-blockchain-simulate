package models

type Block struct {
	ID           int64  `db:"id"`
	BlockNumber  int    `db:"block_number"`
	PreviousHash string `db:"previous_hash"`
	CurrentHash  string `db:"current_hash"`
	CreatedAt    string `db:"created_at"`
}
