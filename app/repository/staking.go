package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type StakingRepository interface {
	Create(ctx context.Context, record models.StakingRecord) (int64, error)
	GetByID(ctx context.Context, id int64) (models.StakingRecord, error)
	GetByUserAddress(ctx context.Context, address string) ([]models.StakingRecord, error)
	GetActiveByUserAddress(ctx context.Context, address string) ([]models.StakingRecord, error)
	GetAllActive(ctx context.Context) ([]models.StakingRecord, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateRewardEarned(ctx context.Context, id int64, amount float64) error
	TotalStakedByAddress(ctx context.Context, address string) (float64, error)
	TotalStakedGlobal(ctx context.Context) (float64, error)
	CreateReward(ctx context.Context, reward models.StakingReward) error
	GetRewardsByAddress(ctx context.Context, address string, limit int) ([]models.StakingReward, error)
}

type stakingRepository struct {
	db *sqlx.DB
}

func NewStakingRepository(db *sqlx.DB) StakingRepository {
	return &stakingRepository{db: db}
}

func (r *stakingRepository) Create(ctx context.Context, record models.StakingRecord) (int64, error) {
	query := `INSERT INTO staking_records (user_address, amount, lock_until, reward_earned, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		record.UserAddress, record.Amount, record.LockUntil,
		record.RewardEarned, record.Status, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *stakingRepository) GetByID(ctx context.Context, id int64) (models.StakingRecord, error) {
	var record models.StakingRecord
	err := r.db.GetContext(ctx, &record,
		`SELECT * FROM staking_records WHERE id = ?`, id)
	return record, err
}

func (r *stakingRepository) GetByUserAddress(ctx context.Context, address string) ([]models.StakingRecord, error) {
	var records []models.StakingRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM staking_records WHERE user_address = ? ORDER BY created_at DESC`, address)
	return records, err
}

func (r *stakingRepository) GetActiveByUserAddress(ctx context.Context, address string) ([]models.StakingRecord, error) {
	var records []models.StakingRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM staking_records WHERE user_address = ? AND status = 'ACTIVE' ORDER BY created_at DESC`, address)
	return records, err
}

func (r *stakingRepository) GetAllActive(ctx context.Context) ([]models.StakingRecord, error) {
	var records []models.StakingRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM staking_records WHERE status = 'ACTIVE'`)
	return records, err
}

func (r *stakingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE staking_records SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *stakingRepository) UpdateRewardEarned(ctx context.Context, id int64, amount float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE staking_records SET reward_earned = reward_earned + ? WHERE id = ?`, amount, id)
	return err
}

func (r *stakingRepository) TotalStakedByAddress(ctx context.Context, address string) (float64, error) {
	var total float64
	err := r.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(amount), 0) FROM staking_records WHERE user_address = ? AND status = 'ACTIVE'`, address)
	return total, err
}

func (r *stakingRepository) TotalStakedGlobal(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(amount), 0) FROM staking_records WHERE status = 'ACTIVE'`)
	return total, err
}

func (r *stakingRepository) CreateReward(ctx context.Context, reward models.StakingReward) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO staking_rewards (stake_id, user_address, reward_amount, block_number, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		reward.StakeID, reward.UserAddress, reward.RewardAmount, reward.BlockNumber, time.Now().Unix())
	return err
}

func (r *stakingRepository) GetRewardsByAddress(ctx context.Context, address string, limit int) ([]models.StakingReward, error) {
	if limit <= 0 {
		limit = 50
	}
	var rewards []models.StakingReward
	err := r.db.SelectContext(ctx, &rewards,
		`SELECT * FROM staking_rewards WHERE user_address = ? ORDER BY created_at DESC LIMIT ?`,
		address, limit)
	return rewards, err
}
