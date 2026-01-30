package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/utils"
	"go.uber.org/zap"
)

type BlockService interface {
	GenerateBlock() (models.Block, error)
	GetBlocks(limit, offset int) ([]models.Block, error)
	GetBlockByID(id int64) (models.Block, error)
	GetBlockByBlockNumber(id int64) (models.Block, error)
	GetDetailsByBlockNumber(id int64) (models.Block, error)
	CheckBlockchainIntegrity() error
}

type blockService struct {
	blockRepo        repository.BlockRepository
	walletRepo       repository.UserWalletRepository
	balanceRepo      repository.UserBalanceRepository
	txRepo           repository.TransactionRepository
	userRepo         repository.UserRepository
	candle           CandleService
	market           MarketEngineService
	publisherWS      *publisher.PublisherWS
	pricingPublisher MarketPricingPublisher
	ledgerPublisher  LedgerPublisher
	rewardPublisher  RewardPublisher
}

func NewBlockService(blockRepo repository.BlockRepository, walletRepo repository.UserWalletRepository, balanceRepo repository.UserBalanceRepository, txRepo repository.TransactionRepository, userRepo repository.UserRepository, candle CandleService, market MarketEngineService, publisherWS *publisher.PublisherWS, pricingPublisher MarketPricingPublisher, ledgerPublisher LedgerPublisher, rewardPublisher RewardPublisher) BlockService {
	return &blockService{
		blockRepo:        blockRepo,
		walletRepo:       walletRepo,
		balanceRepo:      balanceRepo,
		txRepo:           txRepo,
		userRepo:         userRepo,
		candle:           candle,
		market:           market,
		publisherWS:      publisherWS,
		pricingPublisher: pricingPublisher,
		ledgerPublisher:  ledgerPublisher,
		rewardPublisher:  rewardPublisher,
	}
}

func (s *blockService) GenerateBlock() (models.Block, error) {
	// ========================================
	// PHASE 1: Read-only validation (NO LOCKS)
	// ========================================

	// Get last block (read-only)
	lastBlock, err := s.blockRepo.GetLastBlock()
	if err != nil {
		return models.Block{}, fmt.Errorf("get last block: %w", err)
	}

	// get current market state
	var currentMarket models.MarketEngine
	if s.market != nil {
		currentMarket, err = s.market.GetState()
		if err != nil {
			// If error, use default price
			if !strings.Contains(err.Error(), "no rows in result set") {
				return models.Block{}, fmt.Errorf("get market state: %w", err)
			}
			currentMarket.Price = 100.0 // default only on error
		} else if currentMarket.Price == 0 {
			// Fallback if price is 0
			currentMarket.Price = 100.0
		}
	} else {
		// default market state if service not available
		currentMarket.Price = 100.0
	}

	// Get pending transactions (read-only, limit to 100 to prevent timeout)
	pendingTxs, err := s.txRepo.GetPendingTransactions(100)
	if err != nil {
		return models.Block{}, fmt.Errorf("get pending transactions: %w", err)
	}

	if len(pendingTxs) == 0 {
		return models.Block{}, entity.ErrNoPendingTransactions
	}

	var buyVolume, sellVolume float64

	for _, t := range pendingTxs {
		if strings.EqualFold(t.Type, "BUY") {
			buyVolume += t.Amount
		} else if strings.EqualFold(t.Type, "SELL") {
			sellVolume += t.Amount
		}
	}

	// Collect unique addresses
	uniqueAddresses := make(map[string]bool)
	for _, t := range pendingTxs {
		uniqueAddresses[t.FromAddress] = true
		uniqueAddresses[t.ToAddress] = true
	}

	addresses := make([]string, 0, len(uniqueAddresses)+1) // +1 MINER Address
	for addr := range uniqueAddresses {
		addresses = append(addresses, addr)
	}

	// Added miner acount
	addresses = append(addresses, "MINER_ACCOUNT")

	// Get all users at once (read-only)
	users, err := s.userRepo.GetMultipleByAddress(addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("get multiple users: %w", err)
	}

	// Build user cache
	userCache := make(map[string]models.User)
	for _, u := range users {
		userCache[u.Address] = u
	}

	// get all wallets
	wallets, err := s.walletRepo.GetMultipleByAddress(addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("get multiple wallets: %w", err)
	}

	// build wallet cache
	walletCache := make(map[string]models.UserWallet)
	for _, w := range wallets {
		walletCache[w.UserAddress] = w
	}

	// Pre-validate in-memory (no DB)
	balances := make(map[string]float64)
	for _, addr := range addresses {
		wallet, exists := walletCache[addr]

		if exists {
			balances[addr] = wallet.YTEBalance
		} else {
			balances[addr] = 0
		}
	}

	for _, t := range pendingTxs {
		sender, exists := userCache[t.FromAddress]
		if !exists {
			return models.Block{}, fmt.Errorf("sender not found: %s", t.FromAddress)
		}

		// calculato total deduction amount + fee
		totalDeduction := t.Amount + t.Fee

		if balances[sender.Address] < totalDeduction {
			return models.Block{}, fmt.Errorf("insufficient balance for address %s: need %.8f (amount: %.8f + fee: %.8f), have %.8f",
				sender.Address, totalDeduction, t.Amount, t.Fee, balances[sender.Address])
		}

		balances[t.FromAddress] -= totalDeduction // - amount + fee
		balances[t.ToAddress] += t.Amount
	}

	// MINING PHASE : Prof of Work

	// Get all blocks to calculate next difficulty
	allBlocks, err := s.blockRepo.GetAllBlocks()
	if err != nil {
		return models.Block{}, fmt.Errorf("get all blocks: %w", err)
	}

	// calculate difficulty for next block
	difficulty := utils.CalculateNextDifficulty(allBlocks)

	// calculate merkle root
	merkleRoot := utils.CalculateMerkleRoot(pendingTxs)
	logger.LogDebug("Calculated Merkle Root", zap.String("merkle_root", merkleRoot))

	// Calculate block reward
	nextBlockNumber := lastBlock.BlockNumber + 1
	blockReward := utils.CalculateBlockReward(int64(nextBlockNumber))

	// Perform mining (this can take 5-60 seconds depending on difficulty)
	logger.LogInfo("Starting mining process",
		zap.Int64("block_number", int64(nextBlockNumber)),
		zap.Int64("difficulty", int64(difficulty)),
		zap.String("merkle_root", merkleRoot),
		zap.Float64("block_reward", blockReward),
	)

	miningResult := utils.MineBlock(lastBlock.BlockNumber+1, lastBlock.CurrentHash, pendingTxs, difficulty)

	// check if mining was successful
	if miningResult.Hash == "" {
		return models.Block{}, fmt.Errorf("mining failed to find a valid nonce")
	}

	logger.LogInfo("Mining complete", zap.String("hash", miningResult.Hash))

	// ========================================
	// PHASE 2: Write operations (SHORT TRANSACTION)
	// ========================================

	tx, err := s.blockRepo.BeginTx()
	if err != nil {
		return models.Block{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock last block and verify no new block was created
	lastBlockLocked, err := s.blockRepo.GetLastBlockForUpdateWithTx(tx)
	if err != nil {
		return models.Block{}, fmt.Errorf("lock last block: %w", err)
	}

	if lastBlockLocked.BlockNumber != lastBlock.BlockNumber {
		return models.Block{}, fmt.Errorf("new block created while processing, please retry")
	}

	// get locked wallets for update
	lockedWallets, err := s.walletRepo.GetMultipleByAddressWithTx(tx, addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("lock multiple wallets: %w", err)
	}

	// build current wallet cache
	currentWallets := make(map[string]models.UserWallet)
	for _, w := range lockedWallets {
		currentWallets[w.UserAddress] = w
	}

	currentBalances := make(map[string]float64)
	for _, addr := range addresses {
		if wallet, exists := currentWallets[addr]; exists {
			currentBalances[addr] = wallet.YTEBalance
		} else {
			currentBalances[addr] = 0
		}
	}

	// verify miner address
	if _, exists := currentBalances["MINER_ACCOUNT"]; !exists {
		return models.Block{}, fmt.Errorf("MINER_ACCOUNT not found in locked users")
	}

	// Prepare basic info for block creation
	var ledgerEntries []repository.LedgerEntry
	var txIDs []int64
	totalFees := 0.00000000

	// Calculate balances per transaction
	txBalanceChanges := make(map[string]float64)
	for _, t := range pendingTxs {
		totalDeduction := t.Amount + t.Fee
		txBalanceChanges[t.FromAddress] = currentBalances[t.FromAddress] - totalDeduction
		txBalanceChanges[t.ToAddress] = currentBalances[t.ToAddress] + t.Amount
		txBalanceChanges["MINER_ACCOUNT"] = currentBalances["MINER_ACCOUNT"] + t.Fee

		totalFees += t.Fee
		txIDs = append(txIDs, t.ID)
	}

	// Update currentBalances
	for addr, bal := range txBalanceChanges {
		currentBalances[addr] = bal
	}

	// Create block FIRST to get blockID
	newBlock := models.Block{
		BlockNumber:  nextBlockNumber,
		PreviousHash: lastBlock.CurrentHash,
		CurrentHash:  miningResult.Hash,
		Nonce:        miningResult.Nonce,
		Difficulty:   miningResult.Difficulty,
		Timestamp:    time.Now().Unix(),
		MerkleRoot:   merkleRoot,
		MinerAddress: "MINER_ACCOUNT",
		BlockReward:  blockReward,
		TotalFees:    totalFees,
	}

	blockID, err := s.blockRepo.CreateWithTx(tx, newBlock)
	if err != nil {
		return models.Block{}, fmt.Errorf("create block: %w", err)
	}
	newBlock.ID = blockID

	// NOW create ledger entries dengan blockID yang sudah ada
	for _, t := range pendingTxs {
		txID := t.ID
		txIDPtr := &txID
		totalDeduction := t.Amount + t.Fee

		ledgerEntries = append(ledgerEntries,
			repository.LedgerEntry{
				BlockID:      blockID,
				TxID:         txIDPtr,
				Address:      t.FromAddress,
				Amount:       -totalDeduction,
				BalanceAfter: txBalanceChanges[t.FromAddress],
			},
			repository.LedgerEntry{
				BlockID:      blockID,
				TxID:         txIDPtr,
				Address:      t.ToAddress,
				Amount:       t.Amount,
				BalanceAfter: txBalanceChanges[t.ToAddress],
			},
			repository.LedgerEntry{
				BlockID:      blockID,
				TxID:         txIDPtr,
				Address:      "MINER_ACCOUNT",
				Amount:       t.Fee,
				BalanceAfter: txBalanceChanges["MINER_ACCOUNT"],
			},
		)
	}

	// Bulk insert block-transaction links (1 query instead of N)
	err = s.blockRepo.BulkInsertBlockTransactionsWithTx(tx, blockID, txIDs)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk insert block transactions: %w", err)
	}

	// Bulk update transaction status (1 query instead of N)
	err = s.txRepo.BulkMarkConfirmedWithTx(tx, txIDs)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk mark confirmed: %w", err)
	}

	// Bulk update user balances (1 query instead of N)
	walletUpdates := make(map[string]float64)
	for addr, bal := range currentBalances {
		if addr != "MINER_ACCOUNT" {
			walletUpdates[addr] = bal
		}
	}

	err = s.walletRepo.BulkUpdateBalancesWithTx(tx, walletUpdates)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk update balances: %w", err)
	}

	// Process USD Balance Updates For BUY and SELL tx

	var buyerAddresses, sellerAddresses []string
	usdBalances := make(map[string]models.UserBalance)

	for _, t := range pendingTxs {
		if strings.EqualFold(t.Type, "BUY") {
			// Buyer to_address receives YTE Pays USD
			buyerAddresses = append(buyerAddresses, t.ToAddress)
		} else if strings.EqualFold(t.Type, "SELL") {
			// seller : from_address selss YTE receives USD
			sellerAddresses = append(sellerAddresses, t.FromAddress)
		}
	}

	// get all USD with lock
	allUSDAddresses := append(buyerAddresses, sellerAddresses...)
	allUSDAddresses = append(allUSDAddresses, "MINER_ACCOUNT") // miner gets fee

	if len(allUSDAddresses) > 0 {
		lockedUSDBalances, err := s.balanceRepo.GetMultipleByAddressWithTxForUpdate(tx, allUSDAddresses)

		if err != nil {
			return models.Block{}, fmt.Errorf("lock multiple USD balances: %w", err)
		}

		for _, ub := range lockedUSDBalances {
			usdBalances[ub.UserAddress] = ub
		}

		// ensure all address have USD balance record
		for _, addr := range allUSDAddresses {
			if _, exists := usdBalances[addr]; !exists {
				if err := s.balanceRepo.UpsertEmptyIfNotExistsWithTx(tx, addr); err != nil {
					return models.Block{}, fmt.Errorf("upsert empty USD balance: %w", err)
				}

				// refetch after upsert
				balance, err := s.balanceRepo.GetForUpdateWithTx(tx, addr)
				if err != nil {
					return models.Block{}, fmt.Errorf("refetch USD balance after upsert: %w", err)
				}
				usdBalances[addr] = balance
			}
		}
	}

	for _, t := range pendingTxs {
		if strings.EqualFold(t.Type, "BUY") {
			buyerAddr := t.ToAddress
			totalCost := t.Amount + t.Fee

			buyerBalance := usdBalances[buyerAddr]
			balanceAfter := buyerBalance.USDBalance - totalCost

			usdBalances[buyerAddr] = models.UserBalance{
				UserAddress:     buyerAddr,
				USDBalance:      balanceAfter,
				LockedBalance:   buyerBalance.LockedBalance,
				TotalDeposited:  buyerBalance.TotalDeposited,
				TotalWithdrawn:  buyerBalance.TotalWithdrawn + totalCost,
				TotalTraded:     buyerBalance.TotalTraded + t.Amount,
				LastTransaction: buyerBalance.LastTransaction,
			}

			// miner fee receives USD
			minerBalance := usdBalances["MINER_ACCOUNT"]
			minerBalance.USDBalance += t.Fee
			usdBalances["MINER_ACCOUNT"] = minerBalance

		} else if strings.EqualFold(t.Type, "SELL") {
			sellerAddr := t.FromAddress

			// use market price to calculate USD received
			usdAmount := t.Amount * currentMarket.Price
			usdFee := t.Fee * currentMarket.Price

			sellerBalance := usdBalances[sellerAddr]
			balanceAfter := sellerBalance.USDBalance + usdAmount

			usdBalances[sellerAddr] = models.UserBalance{
				UserAddress:     sellerAddr,
				USDBalance:      balanceAfter,
				LockedBalance:   sellerBalance.LockedBalance,
				TotalDeposited:  sellerBalance.TotalDeposited + usdAmount,
				TotalWithdrawn:  sellerBalance.TotalWithdrawn,
				TotalTraded:     sellerBalance.TotalTraded + usdAmount,
				LastTransaction: sellerBalance.LastTransaction,
			}

			// miner fee receives USD
			minerBalance := usdBalances["MINER_ACCOUNT"]
			minerBalance.USDBalance += usdFee
			usdBalances["MINER_ACCOUNT"] = minerBalance
		}
	}

	// bulk update all usd balances
	if err := s.balanceRepo.BulkUpdateBalancesWithTx(tx, usdBalances); err != nil {
		return models.Block{}, fmt.Errorf("bulk update USD balances: %w", err)
	}

	// End process USD

	var marketState models.MarketEngine
	if s.market != nil {
		if marketState, err = s.market.ApplyBlockPricingWithTx(tx, blockID, buyVolume, sellVolume, len(pendingTxs)); err != nil {
			return models.Block{}, fmt.Errorf("apply market pricing: %w", err)
		}
	}

	// store market tick
	var marketTick models.MarketTick
	if s.market != nil {
		marketTick = models.MarketTick{
			BlockID:    blockID,
			Price:      marketState.Price,
			BuyVolume:  buyVolume,
			SellVolume: sellVolume,
			TxCount:    len(pendingTxs),
			CreatedAt:  time.Now().Unix(),
		}
	}

	// Commit (total transaction time: < 2 seconds)
	if err := tx.Commit(); err != nil {
		return models.Block{}, fmt.Errorf("commit transaction: %w", err)
	}

	ctx := context.Background()

	// PHASE 3: Async Event Publishing (POST-COMMIT)

	// publish ledger batch event
	if s.ledgerPublisher != nil {
		ledgerEvents := make([]dto.LedgerEntryEvent, 0, len(ledgerEntries))

		for _, entry := range ledgerEntries {
			event := dto.LedgerEntryEvent{
				Address:      entry.Address,
				Amount:       entry.Amount,
				BalanceAfter: entry.BalanceAfter,
			}
			if entry.TxID != nil {
				event.TxID = entry.TxID
			}

			ledgerEvents = append(ledgerEvents, event)
		}

		if err := s.ledgerPublisher.PublishLedgerBatch(
			ctx,
			blockID,
			newBlock.BlockNumber,
			ledgerEvents,
			newBlock.MinerAddress,
		); err != nil {
			logger.LogWarn("Failed to publish ledger batch", zap.Error(err))
		}
	}

	// publish market pricing event
	if s.pricingPublisher != nil && marketState.ID != 0 {
		if err := s.pricingPublisher.PublishPricingEvent(
			ctx,
			blockID,
			newBlock.BlockNumber,
			marketState,
			marketTick,
			newBlock.MinerAddress,
		); err != nil {
			logger.LogWarn("Failed to publish market pricing event", zap.Error(err))
		}
	}

	//broadcast new block mined
	if s.publisherWS != nil {
		s.publisherWS.Publish(entity.EventTypeBlockMined, newBlock)
	}

	// publish reward calculation event
	if s.rewardPublisher != nil {
		rewardCalcEvent := dto.RewardCalculationEvent{
			BlockID:             blockID,
			BlockNumber:         newBlock.BlockNumber,
			MinerAddress:        newBlock.MinerAddress,
			BlockReward:         newBlock.BlockReward,
			TransactionCount:    len(pendingTxs),
			TotalTransactionFee: newBlock.TotalFees,
			MarketPrice:         marketState.Price,
			Timestamp:           time.Now().Unix(),
		}

		if err := s.rewardPublisher.PublishRewardCalculation(ctx, rewardCalcEvent); err != nil {
			logger.LogWarn("Failed to publish reward calculation event", zap.Error(err))
		}
	}

	// load transactions
	transactions, err := s.txRepo.GetTransactionsByBlockID(blockID)
	if err != nil {
		logger.LogWarn("Failed to load transactions", zap.Error(err))
	} else {
		newBlock.Transactions = transactions
	}

	// send notifycation to websocket
	if s.publisherWS != nil && len(newBlock.Transactions) > 0 {
		for _, tx := range newBlock.Transactions {
			payload := tx
			if tx.FromAddress != "MINER_ACCOUNT" {
				s.publisherWS.PublishToAddress(strings.ToLower(tx.FromAddress), entity.EventTransactionUpdate, payload)
			}

			if tx.ToAddress != "MINER_ACCOUNT" {
				s.publisherWS.PublishToAddress(strings.ToLower(tx.ToAddress), entity.EventTransactionUpdate, payload)
			}
		}
	}

	minerWallet, _ := s.walletRepo.GetByAddress("MINER_ACCOUNT")
	logger.LogBlockEvent(
		int64(newBlock.BlockNumber),
		"mined",
		zap.String("hash", newBlock.CurrentHash),
		zap.String("merkle_root", newBlock.MerkleRoot),
		zap.Int64("nonce", newBlock.Nonce),
		zap.Int64("difficulty", int64(newBlock.Difficulty)),
		zap.Int("transaction_count", len(pendingTxs)),
		zap.Float64("total_fees", totalFees),
		zap.Float64("block_reward", blockReward),
		zap.Float64("total_earned", blockReward+totalFees),
		zap.Float64("miner_balance_before", minerWallet.YTEBalance),
		zap.Float64("current_supply", utils.GetCurrentSupply(int64(nextBlockNumber))),
		zap.Float64("max_supply", utils.GetMaxSupply()),
		zap.Int64("next_halving_block", utils.GetNextHalvingBlock(int64(nextBlockNumber))),
		zap.Int64("blocks_until_halving", utils.GetBlocksUntilHalving(int64(nextBlockNumber))),
		zap.Duration("mining_time", miningResult.Duration),
	)

	return newBlock, nil
}

func (s *blockService) GetBlocks(limit, offset int) ([]models.Block, error) {
	return s.blockRepo.GetBlocks(limit, offset)
}

func (s *blockService) GetBlockByID(id int64) (models.Block, error) {
	return s.blockRepo.GetBlockByID(id)
}

func (s *blockService) GetBlockByBlockNumber(id int64) (models.Block, error) {
	return s.blockRepo.GetBlockByBlockNumber(id)
}

func (s *blockService) CheckBlockchainIntegrity() error {
	blocks, err := s.blockRepo.GetAllBlocks() // ambil semua block

	if err != nil {
		return err
	}

	err = utils.CheckBlockchainIntegrity(blocks)

	return err
}

func (s *blockService) GetDetailsByBlockNumber(id int64) (models.Block, error) {
	var block models.Block

	block, err := s.blockRepo.GetBlockByBlockNumber(id)
	if err != nil {
		return models.Block{}, err
	}

	// Populate transactions for the block
	tx, err := s.txRepo.GetTransactionsByBlockID(block.ID)

	if err != nil {
		return models.Block{}, err
	}

	block.Transactions = tx

	return block, nil
}
