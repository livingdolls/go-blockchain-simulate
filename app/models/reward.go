package models

type RewardInfoResponse struct {
	CurrentBlockNumber int64   `json:"current_block_number"`
	CurrentReward      float64 `json:"current_reward"`
	NextReward         float64 `json:"next_reward"`
	NextHalvingBlock   int64   `json:"next_halving_block"`
	BlocksUntilHalving int64   `json:"blocks_until_halving"`
	CurrentSupply      float64 `json:"current_supply"`
	MaxSupply          float64 `json:"max_supply"`
	SupplyPercentage   float64 `json:"supply_percentage"`
}

type BlockRewardResponse struct {
	BlockNumber  int64   `json:"block_number"`
	MinerAddress string  `json:"miner_address"`
	BlockReward  float64 `json:"block_reward"`
	TotalFees    float64 `json:"total_fees"`
	TotalEarned  float64 `json:"total_earned"`
	Timestamp    int64   `json:"timestamp"`
}

type ScheduleEntry struct {
	BlockNumber int64   `json:"block_number"`
	Reward      float64 `json:"reward"`
	IsHalving   bool    `json:"is_halving"`
}

type ScheduleEntryResponse struct {
	CurrentBlockNumber int64           `json:"current_block_number"`
	Schedule           []ScheduleEntry `json:"schedule"`
}
