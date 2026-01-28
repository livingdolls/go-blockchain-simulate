package dto

type MarketPricingEvent struct {
	BlockID      int64   `json:"block_id"`
	BlockNumber  int     `json:"block_number"`
	Price        float64 `json:"price"`
	Liquidity    float64 `json:"liquidity"`
	BuyVolume    float64 `json:"buy_volume"`
	SellVolume   float64 `json:"sell_volume"`
	TxCount      int     `json:"tx_count"`
	Timestamp    int64   `json:"timestamp"`
	MinerAddress string  `json:"miner_address"`
}

type MarketVolumeUpdate struct {
	BlockID     int64   `json:"block_id"`
	BuyVolume   float64 `json:"buy_volume"`
	SellVolume  float64 `json:"sell_volume"`
	NetVolume   float64 `json:"net_volume"`
	VolumeRatio float64 `json:"volume_ratio"`
	TxCount     int     `json:"tx_count"`
	Timestamp   int64   `json:"timestamp"`
}

type MarketStateSnapshot struct {
	BlockID   int64   `json:"block_id"`
	Price     float64 `json:"price"`
	Liquidity float64 `json:"liquidity"`
	Timestamp int64   `json:"timestamp"`
	CreatedAt int64   `json:"created_at"`
}

type PriceUpdate struct {
	BlockID            int64   `json:"block_id"`
	BlockNumber        int     `json:"block_number"`
	Price              float64 `json:"price"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	Liquidity          float64 `json:"liquidity"`
	BuyVolume          float64 `json:"buy_volume"`
	SellVolume         float64 `json:"sell_volume"`
	TxCount            int     `json:"tx_count"`
	Timestamp          int64   `json:"timestamp"`
	MinerAddress       string  `json:"miner_address"`
}
