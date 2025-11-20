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
}

type blockService struct {
	blockRepo repository.BlockRepository
	txRepo    repository.TransactionRepository
}

func NewBlockService(blockRepo repository.BlockRepository, txRepo repository.TransactionRepository) BlockService {
	return &blockService{
		blockRepo: blockRepo,
		txRepo:    txRepo,
	}
}

func (s *blockService) GenerateBlock() (models.Block, error) {
	pendingTxs, err := s.txRepo.GetPendingTransactions()

	if err != nil {
		return models.Block{}, err
	}

	if len(pendingTxs) == 0 {
		return models.Block{}, fmt.Errorf("no pending transactions")
	}

	// begin tx
	tx, err := s.blockRepo.BeginTx()
	if err != nil {
		return models.Block{}, err
	}

	// ambil hash block terakhir
	lastBlock, err := s.blockRepo.GetLastBlock(tx)

	if err != nil {
		tx.Rollback()
		return models.Block{}, err
	}

	prevHash := lastBlock.CurrentHash
	blockNumber := lastBlock.BlockNumber + 1

	// hash block baru
	newHash := utils.HashBlock(prevHash, pendingTxs)

	block := models.Block{
		BlockNumber:  blockNumber,
		PreviousHash: prevHash,
		CurrentHash:  newHash,
	}

	// save block

	blockID, err := s.blockRepo.CreateWithTx(tx, block)

	if err != nil {
		tx.Rollback()
		return models.Block{}, err
	}

	txIDs := []int64{}
	for _, t := range pendingTxs {
		txIDs = append(txIDs, t.ID)
		err = s.blockRepo.InsertBlockTransactionWithTx(tx, blockID, t.ID)

		if err != nil {
			tx.Rollback()
			return models.Block{}, err
		}
	}

	// update status transaksi jadi confirmed
	err = s.txRepo.MarkConfirmedWithTx(tx, txIDs)
	if err != nil {
		tx.Rollback()
		return models.Block{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.Block{}, err
	}

	block.CreatedAt = time.Now().String()

	return block, nil
}
