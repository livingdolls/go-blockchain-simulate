package models

type Block struct {
	ID           int64         `db:"id" json:"id"`
	BlockNumber  int           `db:"block_number" json:"block_number"`
	PreviousHash string        `db:"previous_hash" json:"previous_hash"`
	CurrentHash  string        `db:"current_hash" json:"current_hash"`
	Nonce        int64         `db:"nonce" json:"nonce"`
	Difficulty   int           `db:"difficulty" json:"difficulty"`
	Timestamp    int64         `db:"timestamp" json:"timestamp"`
	MerkleRoot   string        `db:"merkle_root" json:"merkle_root"`
	CreatedAt    string        `db:"created_at" json:"created_at"`
	Transactions []Transaction `db:"-" json:"transactions,omitempty"`
}
