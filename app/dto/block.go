package dto

import "github.com/livingdolls/go-blockchain-simulate/app/models"

type BlockStatsResponse struct {
	TotalBlocks        int             `json:"total_blocks"`
	AverageBlockTime   float64         `json:"average_block_time"`
	AverageDifficulty  float64         `json:"average_difficulty"`
	TotalTransactions  int             `json:"total_transactions"`
	AvgTxPerBlock      float64         `json:"avg_tx_per_block"`
	TotalBlockRewards  float64         `json:"total_block_rewards"`
	TotalFees          float64         `json:"total_fees"`
	LatestBlock        LatestBlockInfo `json:"latest_block"`
	LastHourBlockCount int64           `json:"last_hour_block_count"`
}

type LatestBlockInfo struct {
	BlockNumber  int64                `json:"block_number"`
	Hash         string               `json:"hash"`
	Timestamp    int64                `json:"timestamp"`
	Transactions []models.Transaction `json:"transactions"`
	MinerAddress string               `json:"miner_address"`
	BlockReward  float64              `json:"block_reward"`
	TotalFees    float64              `json:"total_fees"`
}
