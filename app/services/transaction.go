package services

import (
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type TransactionService interface {
	Send(fromAddress, toAddress, privateKey string, amount float64) (models.Transaction, error)
}

type transactionService struct {
	users   repository.UserRepository
	txs     repository.TransactionRepository
	ledgers repository.LedgerRepository
}

func NewTransactionService(
	users repository.UserRepository,
	txs repository.TransactionRepository,
	ledgers repository.LedgerRepository,
) TransactionService {
	return &transactionService{
		users:   users,
		txs:     txs,
		ledgers: ledgers,
	}
}

func (s *transactionService) Send(fromAddress, toAddress, privateKey string, amount float64) (models.Transaction, error) {
	// first, get sender and receiver
	sender, err := s.users.GetByAddress(fromAddress)
	if err != nil {
		return models.Transaction{}, err
	}

	receiver, err := s.users.GetByAddress(toAddress)

	if err != nil {
		return models.Transaction{}, err
	}

	// validate private key
	if sender.PrivateKey != privateKey {
		return models.Transaction{}, fmt.Errorf("invalid private key for address %s", fromAddress)
	}

	// validate balance
	if sender.Balance < amount {
		return models.Transaction{}, fmt.Errorf("insufficient balance for address %s", fromAddress)
	}

	// create digital signature
	signature := utils.SignFake(privateKey, receiver.Address, amount)
	// create transaction
	tx := models.Transaction{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Signature:   signature,
		Status:      "PENDING",
	}

	txID, err := s.txs.Create(tx)

	if err != nil {
		return models.Transaction{}, err
	}

	tx.ID = txID

	return tx, nil
}
