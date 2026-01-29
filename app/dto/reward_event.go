package dto

type RewardCalculationEvent struct {
	BlockID             int64   `json:"block_id"`
	BlockNumber         int     `json:"block_number"`
	MinerAddress        string  `json:"miner_address"`
	BlockReward         float64 `json:"block_reward"`
	TransactionCount    int     `json:"transaction_count"`
	TotalTransactionFee float64 `json:"total_transaction_fee"`
	MarketPrice         float64 `json:"market_price"`
	Timestamp           int64   `json:"timestamp"`
}

type RewardDistributionEvent struct {
	BlockID         int64           `json:"block_id"`
	BlockNumber     int             `json:"block_number"`
	MinerAddress    string          `json:"miner_address"`
	MinerReward     float64         `json:"miner_reward"`
	MinerUSDValue   float64         `json:"miner_usd_value"`
	RewardBreakdown RewardBreakDown `json:"reward_breakdown"`
	Timestamp       int64           `json:"timestamp"`
}

type RewardBreakDown struct {
	BlockReward       float64 `json:"block_reward"`
	TransactionFees   float64 `json:"transaction_fees"`
	BonusReward       float64 `json:"bonus_reward"`
	TotalReward       float64 `json:"total_reward"`
	EstimatedUSDValue float64 `json:"estimated_usd_value"`
}

type RewardDistributionResult struct {
	Status           string  `json:"status"`
	BlockNumber      int     `json:"block_number"`
	MinerAddress     string  `json:"miner_address"`
	TotalRewardGiven float64 `json:"total_reward_given"`
	DistributedAt    int64   `json:"distributed_at"`
	TransactionHash  string  `json:"transaction_hash"`
	Error            string  `json:"error,omitempty"`
}
