package services

import (
	"fmt"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type BlockService interface {
	GenerateBlock() (models.Block, error)
	GetBlocks(limit, offset int) ([]models.Block, error)
	GetBlockByID(id int64) (models.Block, error)
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

	addresses := make([]string, 0, len(uniqueAddresses))
	for addr := range uniqueAddresses {
		addresses = append(addresses, addr)
	}

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

		if balances[sender.Address] < t.Amount {
			return models.Block{}, fmt.Errorf("insufficient balance for address %s", sender.Address)
		}

		balances[t.FromAddress] -= t.Amount
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

	// Perform mining (this can take 5-60 seconds depending on difficulty)

	fmt.Printf("Starting mining process...\n")
	miningResult := utils.MineBlock(lastBlock.BlockNumber+1, lastBlock.CurrentHash, pendingTxs, difficulty)

	// check if mining was successful
	if miningResult.Hash == "" {
		return models.Block{}, fmt.Errorf("mining failed to find a valid nonce")
	}

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

	// Create block
	newBlock := models.Block{
		BlockNumber:  lastBlock.BlockNumber + 1,
		PreviousHash: lastBlock.CurrentHash,
		CurrentHash:  miningResult.Hash,
		Nonce:        miningResult.Nonce,
		Difficulty:   miningResult.Difficulty,
		Timestamp:    time.Now().Unix(),
		MerkleRoot:   merkleRoot,
	}

	blockID, err := s.blockRepo.CreateWithTx(tx, newBlock)
	if err != nil {
		return models.Block{}, fmt.Errorf("create block: %w", err)
	}
	newBlock.ID = blockID

	// Lock all users at once
	err = s.userRepo.LockMultipleUsersWithTx(tx, addresses)
	if err != nil {
		return models.Block{}, fmt.Errorf("lock users: %w", err)
	}

	// Prepare bulk operations
	var ledgerEntries []repository.LedgerEntry
	var txIDs []int64

	currentBalances := make(map[string]float64)
	for _, u := range users {
		currentBalances[u.Address] = u.Balance
	}

	for _, t := range pendingTxs {
		// Calculate balances
		currentBalances[t.FromAddress] -= t.Amount
		currentBalances[t.ToAddress] += t.Amount

		// Prepare ledger entries
		ledgerEntries = append(ledgerEntries,
			repository.LedgerEntry{
				TxID:         t.ID,
				Address:      t.FromAddress,
				Amount:       -t.Amount,
				BalanceAfter: currentBalances[t.FromAddress],
			},
			repository.LedgerEntry{
				TxID:         t.ID,
				Address:      t.ToAddress,
				Amount:       t.Amount,
				BalanceAfter: currentBalances[t.ToAddress],
			},
		)

		txIDs = append(txIDs, t.ID)
	}

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

	fmt.Printf("\n✅ Block #%d added to blockchain!\n", newBlock.BlockNumber)
	fmt.Printf("   Hash: %s\n", newBlock.CurrentHash)
	fmt.Printf("   Nonce: %d\n", newBlock.Nonce)
	fmt.Printf("   Difficulty: %d\n", newBlock.Difficulty)
	fmt.Printf("   Transactions: %d\n", len(pendingTxs))
	fmt.Printf("   Mining time: %v\n\n", miningResult.Duration)

	return newBlock, nil
}

func (s *blockService) GetBlocks(limit, offset int) ([]models.Block, error) {
	return s.blockRepo.GetBlocks(limit, offset)
}

func (s *blockService) GetBlockByID(id int64) (models.Block, error) {
	return s.blockRepo.GetBlockByID(id)
}

func (s *blockService) CheckBlockchainIntegrity() error {
	blocks, err := s.blockRepo.GetAllBlocks() // ambil semua block

	if err != nil {
		return err
	}

	err = utils.CheckBlockchainIntegrity(blocks)

	return err
}
