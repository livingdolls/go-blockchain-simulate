package services

import (
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type RewardService interface {
	RewardInfo() (models.RewardInfoResponse, error)
	GetRewardSchedule(block int) (models.ScheduleEntryResponse, error)
}

type rewardService struct {
	blockRepo repository.BlockRepository
}

func NewRewardHandler(blockRepo repository.BlockRepository) RewardService {
	return &rewardService{
		blockRepo: blockRepo,
	}
}

// RewardInfo implements RewardService.
func (r *rewardService) RewardInfo() (models.RewardInfoResponse, error) {
	lastBlock, err := r.blockRepo.GetLastBlock()

	if err != nil {
		return models.RewardInfoResponse{}, err
	}

	currentReward := utils.CalculateBlockReward(int64(lastBlock.BlockNumber))
	nextReward := utils.CalculateBlockReward(int64(lastBlock.BlockNumber) + 1)
	currentSupply := utils.GetCurrentSupply(int64(lastBlock.BlockNumber))
	maxSupply := utils.GetMaxSupply()

	return models.RewardInfoResponse{
		CurrentBlockNumber: int64(lastBlock.BlockNumber),
		CurrentReward:      currentReward,
		NextReward:         nextReward,
		NextHalvingBlock:   utils.GetNextHalvingBlock(int64(lastBlock.BlockNumber)),
		BlocksUntilHalving: utils.GetBlocksUntilHalving(int64(lastBlock.BlockNumber)),
		CurrentSupply:      currentSupply,
		MaxSupply:          maxSupply,
		SupplyPercentage:   (currentSupply / maxSupply) * 100,
	}, nil
}

func (r *rewardService) GetRewardSchedule(block int) (models.ScheduleEntryResponse, error) {
	lastBlock, err := r.blockRepo.GetLastBlock()

	if err != nil {
		return models.ScheduleEntryResponse{}, err
	}

	schedule := make([]models.ScheduleEntry, 0, block)

	for i := 1; i <= block; i++ {
		blockNumber := int64(lastBlock.BlockNumber) + int64(i)
		reward := utils.CalculateBlockReward(blockNumber)
		isHalving := blockNumber%utils.HalvingInterval == 0

		schedule = append(schedule, models.ScheduleEntry{
			BlockNumber: blockNumber,
			Reward:      reward,
			IsHalving:   isHalving,
		})
	}

	return models.ScheduleEntryResponse{
		CurrentBlockNumber: int64(lastBlock.BlockNumber),
		Schedule:           schedule,
	}, err
}
