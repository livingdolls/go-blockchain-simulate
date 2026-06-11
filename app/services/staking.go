package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"go.uber.org/zap"
)

type StakingService interface {
	Stake(ctx context.Context, address string, amount float64, lockDuration int64) (models.StakingRecord, error)
	Unstake(ctx context.Context, address string, stakeID int64) error
	GetStakingStatus(ctx context.Context, address string) (models.StakingStatus, error)
	GetStakingInfo(ctx context.Context) (models.StakingInfo, error)
	DistributeRewards(ctx context.Context, blockNumber int64) error
}

type stakingService struct {
	stakingRepo   repository.StakingRepository
	walletRepo    repository.UserWalletRepository
	blockRepo     repository.BlockRepository
	notifPublisher publisher.NotificationPublisher
}

func NewStakingService(
	stakingRepo repository.StakingRepository,
	walletRepo repository.UserWalletRepository,
	blockRepo repository.BlockRepository,
	notifPublisher publisher.NotificationPublisher,
) StakingService {
	return &stakingService{
		stakingRepo:    stakingRepo,
		walletRepo:     walletRepo,
		blockRepo:      blockRepo,
		notifPublisher: notifPublisher,
	}
}

func (s *stakingService) Stake(ctx context.Context, address string, amount float64, lockDuration int64) (models.StakingRecord, error) {
	address = strings.ToLower(strings.TrimSpace(address))

	// Validasi amount
	if amount < models.DefaultMinStakeAmount {
		return models.StakingRecord{}, fmt.Errorf("minimum stake amount is %.2f YTE", models.DefaultMinStakeAmount)
	}

	// Validasi lock duration
	if lockDuration < models.DefaultMinLockDuration {
		return models.StakingRecord{}, fmt.Errorf("minimum lock duration is %d seconds", models.DefaultMinLockDuration)
	}

	// Cek balance user
	wallet, err := s.walletRepo.GetByAddress(address)
	if err != nil {
		return models.StakingRecord{}, fmt.Errorf("wallet not found: %w", err)
	}

	if wallet.YTEBalance < amount {
		return models.StakingRecord{}, fmt.Errorf("insufficient balance: have %.8f, need %.8f", wallet.YTEBalance, amount)
	}

	// Kurangi balance user
	newBalance := wallet.YTEBalance - amount
	if err := s.walletRepo.UpdateWalletWithTx(nil, address, newBalance); err != nil {
		return models.StakingRecord{}, fmt.Errorf("failed to deduct balance: %w", err)
	}

	// Buat staking record
	record := models.StakingRecord{
		UserAddress:  address,
		Amount:       amount,
		LockUntil:    time.Now().Unix() + lockDuration,
		RewardEarned: 0,
		Status:       models.StakingStatusActive,
	}

	id, err := s.stakingRepo.Create(ctx, record)
	if err != nil {
		// Rollback balance jika create gagal
		_ = s.walletRepo.UpdateWalletWithTx(nil, address, wallet.YTEBalance)
		return models.StakingRecord{}, fmt.Errorf("failed to create staking record: %w", err)
	}

	record.ID = id

	// Publish notification
	if s.notifPublisher != nil {
		notif := dto.NewNotificationEvent(
			dto.TypeBalanceUpdated,
			dto.PriorityMedium,
			address,
			"Staking Berhasil",
			fmt.Sprintf("Anda telah melakukan staking %.8f YTE selama %d hari", amount, lockDuration/86400),
			[]dto.NotificationChannel{dto.ChannelWebSocket},
		)
		notif.Data = map[string]interface{}{
			"action":         "stake",
			"amount":         amount,
			"lock_until":     record.LockUntil,
			"stake_id":       id,
		}
		_ = s.notifPublisher.Publish(ctx, *notif)
	}

	logger.LogInfo("Staking created",
		zap.String("address", address),
		zap.Float64("amount", amount),
		zap.Int64("lock_until", record.LockUntil),
	)

	return record, nil
}

func (s *stakingService) Unstake(ctx context.Context, address string, stakeID int64) error {
	address = strings.ToLower(strings.TrimSpace(address))

	record, err := s.stakingRepo.GetByID(ctx, stakeID)
	if err != nil {
		return fmt.Errorf("stake not found: %w", err)
	}

	if record.UserAddress != address {
		return fmt.Errorf("%w: stake belongs to different address", entity.ErrUnauthorized)
	}

	if record.Status != models.StakingStatusActive {
		return fmt.Errorf("stake is not active (status: %s)", record.Status)
	}

	if !record.IsUnlocked() {
		return fmt.Errorf("stake is still locked for %d days", record.DaysRemaining())
	}

	// Update status
	if err := s.stakingRepo.UpdateStatus(ctx, stakeID, models.StakingStatusUnstaking); err != nil {
		return fmt.Errorf("failed to update stake status: %w", err)
	}

	// Kembalikan balance + reward
	wallet, err := s.walletRepo.GetByAddress(address)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	totalReturn := record.Amount + record.RewardEarned
	newBalance := wallet.YTEBalance + totalReturn
	if err := s.walletRepo.UpdateWalletWithTx(nil, address, newBalance); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	// Mark as withdrawn
	if err := s.stakingRepo.UpdateStatus(ctx, stakeID, models.StakingStatusWithdrawn); err != nil {
		return fmt.Errorf("failed to mark as withdrawn: %w", err)
	}

	// Publish notification
	if s.notifPublisher != nil {
		notif := dto.NewNotificationEvent(
			dto.TypeBalanceUpdated,
			dto.PriorityHigh,
			address,
			"Unstake Berhasil",
			fmt.Sprintf("Anda telah menarik %.8f YTE (termasuk reward %.8f YTE)", totalReturn, record.RewardEarned),
			[]dto.NotificationChannel{dto.ChannelWebSocket},
		)
		notif.Data = map[string]interface{}{
			"action":      "unstake",
			"amount":      record.Amount,
			"reward":      record.RewardEarned,
			"total":       totalReturn,
			"stake_id":    stakeID,
		}
		_ = s.notifPublisher.Publish(ctx, *notif)
	}

	logger.LogInfo("Staking unstaked",
		zap.String("address", address),
		zap.Int64("stake_id", stakeID),
		zap.Float64("amount", record.Amount),
		zap.Float64("reward", record.RewardEarned),
	)

	return nil
}

func (s *stakingService) GetStakingStatus(ctx context.Context, address string) (models.StakingStatus, error) {
	address = strings.ToLower(strings.TrimSpace(address))

	records, err := s.stakingRepo.GetByUserAddress(ctx, address)
	if err != nil {
		return models.StakingStatus{}, fmt.Errorf("failed to get staking records: %w", err)
	}

	totalStaked, _ := s.stakingRepo.TotalStakedByAddress(ctx, address)

	var totalRewards float64
	activeCount := 0
	for _, r := range records {
		totalRewards += r.RewardEarned
		if r.Status == models.StakingStatusActive {
			activeCount++
		}
	}

	return models.StakingStatus{
		TotalStaked:  totalStaked,
		TotalRewards: totalRewards,
		ActiveStakes: activeCount,
		Records:      records,
	}, nil
}

func (s *stakingService) GetStakingInfo(ctx context.Context) (models.StakingInfo, error) {
	totalStaked, _ := s.stakingRepo.TotalStakedGlobal(ctx)

	return models.StakingInfo{
		TotalStaked:     totalStaked,
		StakingAPR:      models.DefaultStakingAPR,
		MinStakeAmount:  models.DefaultMinStakeAmount,
		MinLockDuration: models.DefaultMinLockDuration,
		RewardPerBlock:  calculateRewardPerBlock(totalStaked),
	}, nil
}

func (s *stakingService) DistributeRewards(ctx context.Context, blockNumber int64) error {
	records, err := s.stakingRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active stakes: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	totalStaked, _ := s.stakingRepo.TotalStakedGlobal(ctx)
	if totalStaked == 0 {
		return nil
	}

	// Hitung total reward pool untuk block ini (misal 0.1% dari total staked per block)
	rewardPool := totalStaked * 0.001 // 0.1% per block distribution

	for _, record := range records {
		// Proportional reward: (stake_amount / total_staked) * reward_pool
		proportion := record.Amount / totalStaked
		reward := proportion * rewardPool

		if reward <= 0 {
			continue
		}

		// Update reward earned
		if err := s.stakingRepo.UpdateRewardEarned(ctx, record.ID, reward); err != nil {
			logger.LogError("failed to update staking reward", err,
				zap.Int64("stake_id", record.ID),
			)
			continue
		}

		// Log reward
		stakingReward := models.StakingReward{
			StakeID:      record.ID,
			UserAddress:  record.UserAddress,
			RewardAmount: reward,
			BlockNumber:  blockNumber,
		}
		if err := s.stakingRepo.CreateReward(ctx, stakingReward); err != nil {
			logger.LogError("failed to create staking reward log", err)
		}

		// Publish notification setiap 10 blocks (mengurangi noise)
		if blockNumber%10 == 0 && s.notifPublisher != nil {
			notif := dto.NewNotificationEvent(
				dto.TypeRewardEarned,
				dto.PriorityLow,
				record.UserAddress,
				"Staking Reward",
				fmt.Sprintf("Anda mendapat reward %.8f YTE dari staking", reward),
				[]dto.NotificationChannel{dto.ChannelWebSocket},
			)
			notif.Data = map[string]interface{}{
				"action":      "staking_reward",
				"stake_id":    record.ID,
				"reward":      reward,
				"block_number": blockNumber,
			}
			_ = s.notifPublisher.Publish(ctx, *notif)
		}
	}

	logger.LogInfo("Staking rewards distributed",
		zap.Int64("block", blockNumber),
		zap.Int("stakes", len(records)),
		zap.Float64("reward_pool", rewardPool),
	)

	return nil
}

// calculateRewardPerBlock menghitung reward per block berdasarkan total staked.
// Formula: (totalStaked * APR / 365 / 24 / 60 / 10) dimana 10 = blocks per menit
func calculateRewardPerBlock(totalStaked float64) float64 {
	if totalStaked == 0 {
		return 0
	}
	blocksPerYear := float64(365 * 24 * 60 * 10) // ~5.2M blocks per tahun
	return totalStaked * (models.DefaultStakingAPR / 100) / blocksPerYear
}

// unused suppressor
var _ = math.MaxInt64
