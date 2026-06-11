package models

import "time"

// StakingRecord mewakili satu stake oleh user.
type StakingRecord struct {
	ID           int64   `db:"id" json:"id"`
	UserAddress  string  `db:"user_address" json:"user_address"`
	Amount       float64 `db:"amount" json:"amount"`
	LockUntil    int64   `db:"lock_until" json:"lock_until"`
	RewardEarned float64 `db:"reward_earned" json:"reward_earned"`
	Status       string  `db:"status" json:"status"` // ACTIVE, UNSTAKING, WITHDRAWN
	CreatedAt    string  `db:"created_at" json:"created_at"`
}

// StakingReward mewakili satu reward distribution entry.
type StakingReward struct {
	ID           int64   `db:"id" json:"id"`
	StakeID      int64   `db:"stake_id" json:"stake_id"`
	UserAddress  string  `db:"user_address" json:"user_address"`
	RewardAmount float64 `db:"reward_amount" json:"reward_amount"`
	BlockNumber  int64   `db:"block_number" json:"block_number"`
	CreatedAt    string  `db:"created_at" json:"created_at"`
}

// StakingStatus response untuk GET /staking/status.
type StakingStatus struct {
	TotalStaked   float64         `json:"total_staked"`
	TotalRewards  float64         `json:"total_rewards"`
	ActiveStakes  int             `json:"active_stakes"`
	Records       []StakingRecord `json:"records"`
}

// StakingInfo response untuk GET /staking/info.
type StakingInfo struct {
	TotalStaked      float64 `json:"total_staked"`
	StakingAPR       float64 `json:"staking_apr"`
	MinStakeAmount   float64 `json:"min_stake_amount"`
	MinLockDuration  int64   `json:"min_lock_duration_seconds"`
	RewardPerBlock   float64 `json:"reward_per_block"`
	NextRewardBlock  int64   `json:"next_reward_block"`
}

// Status constants
const (
	StakingStatusActive     = "ACTIVE"
	StakingStatusUnstaking = "UNSTAKING"
	StakingStatusWithdrawn = "WITHDRAWN"
)

// Default staking config
const (
	DefaultMinStakeAmount  = 1.0     // minimum 1 YTE
	DefaultMinLockDuration = 86400   // minimum 1 hari (detik)
	DefaultStakingAPR      = 10.0    // 10% APR
)

// IsUnlocked apakah stake sudah bisa di-unstake?
func (s *StakingRecord) IsUnlocked() bool {
	return time.Now().Unix() >= s.LockUntil
}

// DaysRemaining berapa hari lagi sebelum unlock.
func (s *StakingRecord) DaysRemaining() int {
	remaining := s.LockUntil - time.Now().Unix()
	if remaining <= 0 {
		return 0
	}
	return int(remaining / 86400)
}
