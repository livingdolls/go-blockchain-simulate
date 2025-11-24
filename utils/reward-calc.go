package utils

import "math"

const (
	// initial block = 50 (like btc)
	InitialBlockReward = 50

	// Halving interval (reduce reward every N blocks)
	// Bitcoin : 210,000 blocks (4 years)
	HalvingInterval = 100 // For Demo

	// Minimum reward (never go below this)
	MinimumReward = 0.00000001
)

func CalculateBlockReward(blockNumber int64) float64 {
	// calculate how many halvings have occurred
	// example : 250 / 100 = 2
	halvings := blockNumber / HalvingInterval

	// calculate reward: baseReward / (2^halvings)
	// example 50 / 2^2 = 50 / 4 = 12.5
	reward := InitialBlockReward / math.Pow(2, float64(halvings))

	// ensure reward doesn't go below minimum
	if reward < MinimumReward {
		reward = MinimumReward
	}

	return reward
}

// get current supply, calculate total coins minted up to a block number
// how to works:
// example blockNumber = 250
// period 1 -> block 1-100 reward 50
// period 2 -> block 101-200 reward 25
// period 3 -> block 201-250 reward 12.5
// total supply 100*50 + 100*25 + 50*12.5 = 5000 + 250 + 625 = 8125
func GetCurrentSupply(blockNumber int64) float64 {
	if blockNumber <= 0 {
		return 0
	}

	totalSupply := 0.0

	// calculate supply for each halving period
	currentBlock := int64(1)

	for currentBlock <= blockNumber {
		// calculate end of current halving period
		// example currentBlock = 150
		// (150-1)/ 100 = 1
		// 1+1 = 2
		// 2*100 = 200 -> end period
		endBlock := ((currentBlock-1)/HalvingInterval + 1) * HalvingInterval

		if endBlock > blockNumber {
			endBlock = blockNumber
		}

		// Calculate blocks in this period
		blocksInPeriod := endBlock - currentBlock + 1

		// calculate reward for this period
		reward := CalculateBlockReward(currentBlock)

		// add to total supply
		totalSupply += float64(blocksInPeriod) * reward

		// move to next period
		currentBlock = endBlock + 1
	}

	return totalSupply
}

func GetMaxSupply() float64 {
	// calculate supply for a every large number of blocks
	// after 20 halvings, reward becomes essentially zero
	maxHalvings := 20 //real bitcoin is 33
	totalSupply := 0.0

	//count max supply
	// example :
	// Halving 0 -> reward: 50 -> supply 100*50 = 5000
	// Halving 1 -> reward: 25 -> supply 2500
	// halving 2 -> reward: 12.5 -> 1250
	for i := 0; i < maxHalvings; i++ {
		blocksInPeriod := int64(HalvingInterval)
		reward := InitialBlockReward / math.Pow(2, float64(i))
		totalSupply += float64(blocksInPeriod) * reward
	}

	return totalSupply
}

func GetNextHalvingBlock(currentBlock int64) int64 {
	// example currentBlock = 250
	// (250/100 +1)*100 = 300
	return ((currentBlock / HalvingInterval) + 1) * HalvingInterval
}

func GetBlocksUntilHalving(currentBlock int64) int64 {
	//example currentBlock 250 -> 300 - 250 = 50 block
	nextHalving := GetNextHalvingBlock(currentBlock)
	return nextHalving - currentBlock
}
