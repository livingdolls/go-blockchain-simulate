package models

type MarketEngine struct {
	ID        int     `db:"id" json:"id"`
	Price     float64 `db:"price" json:"price"`
	Liquidity float64 `db:"liquidity" json:"liquidity"`
	LastBlock int64   `db:"last_block" json:"last_block"`
	UpdatedAt string  `db:"updated_at" json:"updated_at"`
}

type MarketTick struct {
	ID         int     `db:"id" json:"id"`
	BlockID    int64   `db:"block_id" json:"block_id"`
	Price      float64 `db:"price" json:"price"`
	BuyVolume  float64 `db:"buy_volume" json:"buy_volume"`
	SellVolume float64 `db:"sell_volume" json:"sell_volume"`
	TxCount    int     `db:"tx_count" json:"tx_count"`
	CreatedAt  int64   `db:"created_at" json:"created_at"`
}
