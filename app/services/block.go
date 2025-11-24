package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type BlockService interface {
	GenerateBlock() (models.Block, error)
	GetBlocks(limit, offset int) ([]models.Block, error)
	GetBlockByID(id int64) (models.Block, error)
	GetBlockByBlockNumber(id int64) (models.Block, error)
	CheckBlockchainIntegrity() error
}

type blockService struct {
	blockRepo  repository.BlockRepository
	txRepo     repository.TransactionRepository
	userRepo   repository.UserRepository
	ledgerRepo repository.LedgerRepository
}

func NewBlockService(blockRepo repository.BlockRepository, txRepo repository.TransactionRepository, userRepo repository.UserRepository, ledgerRepo repository.LedgerRepository) BlockService {
	return &blockService{
		blockRepo:  blockRepo,
		txRepo:     txRepo,
		userRepo:   userRepo,
		ledgerRepo: ledgerRepo,
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

	// Get pending transactions (read-only, limit to 100 to prevent timeout)
	pendingTxs, err := s.txRepo.GetPendingTransactions(100)
	if err != nil {
		return models.Block{}, fmt.Errorf("get pending transactions: %w", err)
	}

	if len(pendingTxs) == 0 {
		return models.Block{}, fmt.Errorf("no pending transactions")
	}

	// Collect unique addresses
	uniqueAddresses := make(map[string]bool)
	for _, t := range pendingTxs {
		uniqueAddresses[t.FromAddress] = true
		uniqueAddresses[t.ToAddress] = true
	}

	uniqueAddresses["FEE_POOL"] = true

	addresses := make([]string, 0, len(uniqueAddresses)+2) // +2 MINER Address and FEE_POOL
	for addr := range uniqueAddresses {
		addresses = append(addresses, addr)
	}

	// Added fee pool
	addresses = append(addresses, "FEE_POOL")
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

	// Pre-validate in-memory (no DB)
	balances := make(map[string]float64)
	for _, u := range users {
		balances[u.Address] = u.Balance
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
	fmt.Printf("Calculated Merkle Root: %s\n", merkleRoot)

	// Calculate block reward
	nextBlockNumber := lastBlock.BlockNumber + 1
	blockReward := utils.CalculateBlockReward(int64(nextBlockNumber))

	// Perform mining (this can take 5-60 seconds depending on difficulty)
	fmt.Printf("Starting mining process...\n")
	fmt.Printf("Block Number: %d\n", nextBlockNumber)
	fmt.Printf("Difficulty: %d\n", difficulty)
	fmt.Printf("Merkle Root: %s\n", merkleRoot)
	fmt.Printf("Block Reward: %.8f\n", blockReward)

	miningResult := utils.MineBlock(lastBlock.BlockNumber+1, lastBlock.CurrentHash, pendingTxs, difficulty)

	// check if mining was successful
	if miningResult.Hash == "" {
		return models.Block{}, fmt.Errorf("mining failed to find a valid nonce")
	}

	fmt.Printf("Mining complete! hash: %s\n", miningResult.Hash)

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

	currentBalances := make(map[string]float64)
	for _, u := range users {
		currentBalances[u.Address] = u.Balance
	}

	// verify fee
	if _, exists := currentBalances["FEE_POOL"]; !exists {
		return models.Block{}, fmt.Errorf("FEE_POOL not found in locked users")
	}

	// verify miner address
	if _, exists := currentBalances["MINER_ACCOUNT"]; !exists {
		return models.Block{}, fmt.Errorf("MINER_ACCOUNT not found in locked users")
	}

	// Prepare bulk operations
	var ledgerEntries []repository.LedgerEntry
	var txIDs []int64
	totalFees := 0.00000000

	for _, t := range pendingTxs {
		txID := t.ID // int64
		txIDPtr := &txID
		// Calculate balances
		totalDeduction := t.Amount + t.Fee
		currentBalances[t.ToAddress] += t.Amount
		currentBalances[t.FromAddress] -= totalDeduction

		currentBalances["MINER_ACCOUNT"] += t.Fee
		totalFees += t.Fee

		// Prepare ledger entries
		ledgerEntries = append(ledgerEntries,
			repository.LedgerEntry{
				TxID:         txIDPtr,
				Address:      t.FromAddress,
				Amount:       -totalDeduction,
				BalanceAfter: currentBalances[t.FromAddress],
			},
			repository.LedgerEntry{
				TxID:         txIDPtr,
				Address:      t.ToAddress,
				Amount:       t.Amount,
				BalanceAfter: currentBalances[t.ToAddress],
			},
			repository.LedgerEntry{
				TxID:         txIDPtr,
				Address:      "MINER_ACCOUNT",
				Amount:       t.Fee,
				BalanceAfter: currentBalances["MINER_ACCOUNT"],
			},
		)

		txIDs = append(txIDs, t.ID)
	}

	// Add block reward to miner
	currentBalances["MINER_ACCOUNT"] += blockReward

	// create ledger entry for block reward
	ledgerEntries = append(ledgerEntries, repository.LedgerEntry{
		TxID:         nil, // 0 means coinbase/block reward
		Address:      "MINER_ACCOUNT",
		Amount:       blockReward,
		BalanceAfter: currentBalances["MINER_ACCOUNT"],
	})

	// Create block
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

	// Bulk insert ledger entries (1 query instead of N*2)
	err = s.ledgerRepo.BulkCreateWithTx(tx, ledgerEntries)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk create ledger: %w", err)
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
	err = s.userRepo.BulkUpdateBalancesWithTx(tx, currentBalances)
	if err != nil {
		return models.Block{}, fmt.Errorf("bulk update balances: %w", err)
	}

	// Commit (total transaction time: < 2 seconds)
	if err := tx.Commit(); err != nil {
		return models.Block{}, fmt.Errorf("commit transaction: %w", err)
	}

	// load transactions
	newBlock.Transactions, _ = s.txRepo.GetTransactionsByBlockID(blockID)
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("✅ BLOCK #%d SUCCESSFULLY MINED\n", newBlock.BlockNumber)
	fmt.Printf(strings.Repeat("=", 60) + "\n\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n\n")

	fmt.Printf("📦 Block Information:\n")
	fmt.Printf("   Hash:             %s\n", newBlock.CurrentHash)
	fmt.Printf("   Merkle Root:      %s\n", newBlock.MerkleRoot)
	fmt.Printf("   Nonce:            %d\n", newBlock.Nonce)
	fmt.Printf("   Difficulty:       %d\n", newBlock.Difficulty)

	fmt.Printf("\n💰 Transaction Summary:\n")
	fmt.Printf("   Transactions:     %d\n", len(pendingTxs))
	fmt.Printf("   Total Fees:       %.8f\n", totalFees)

	fmt.Printf("\n🏆 Mining Reward:\n")
	fmt.Printf("   Block Reward:     %.8f\n", blockReward)
	fmt.Printf("   Transaction Fees: %.8f\n", totalFees)
	fmt.Printf("   Total Earned:     %.8f\n", blockReward+totalFees)
	fmt.Printf("   Miner Balance:    %.8f\n", currentBalances["MINER_ACCOUNT"])

	fmt.Printf("\n📊 Network Statistics:\n")
	fmt.Printf("   Current Supply:   %.8f\n", utils.GetCurrentSupply(int64(nextBlockNumber)))
	fmt.Printf("   Max Supply:       %.8f\n", utils.GetMaxSupply())
	fmt.Printf("   Next Halving:     Block #%d\n", utils.GetNextHalvingBlock(int64(nextBlockNumber)))
	fmt.Printf("   Blocks Until:     %d\n", utils.GetBlocksUntilHalving(int64(nextBlockNumber)))

	fmt.Printf("\n⛏️  Mining Performance:\n")
	fmt.Printf("   Mining Time:      %v\n", miningResult.Duration)

	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n\n")

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
