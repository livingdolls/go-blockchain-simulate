package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
)

type RewardPublisher interface {
	PublishRewardCalculation(ctx context.Context, event dto.RewardCalculationEvent) error
	PublishRewardDistribution(ctx context.Context, event dto.RewardDistributionEvent) error
}

type rewardPublisher struct {
	client *rabbitmq.Client
}

func NewRewardPublisher(client *rabbitmq.Client) RewardPublisher {
	return &rewardPublisher{client: client}
}

func (rp *rewardPublisher) PublishRewardCalculation(ctx context.Context, event dto.RewardCalculationEvent) error {
	eventJSON, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("failed to marshal reward calculation event: %v", err)
	}

	err = rp.client.Publish(
		ctx,
		rabbitmq.RewardExchange,
		rabbitmq.RewardCalculationKey,
		eventJSON,
	)

	if err != nil {
		log.Printf("[REWARD_PUBLISHER] Failed to publish reward calculation event: %v", err)
		return err
	}

	log.Printf("[REWARD_PUBLISHER] Published reward calculation event for block %d", event.BlockNumber)

	return nil
}

func (rp *rewardPublisher) PublishRewardDistribution(ctx context.Context, event dto.RewardDistributionEvent) error {
	eventJSON, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("failed to marshal reward distribution event: %v", err)
	}

	err = rp.client.Publish(
		ctx,
		rabbitmq.RewardExchange,
		rabbitmq.RewardDistributionKey,
		eventJSON,
	)

	if err != nil {
		log.Printf("[REWARD_PUBLISHER] Failed to publish reward distribution event: %v", err)
		return err
	}

	log.Printf("[REWARD_PUBLISHER] Published reward distribution event for block %d", event.BlockNumber)

	return nil
}
