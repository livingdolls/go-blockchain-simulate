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

	// start database transaction
	trx, err := s.users.BeginTx()
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer trx.Rollback()

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

	txID, err := s.txs.CreateWithTx(trx, tx)

	if err != nil {
		return models.Transaction{}, err
	}

	// update balances
	newSenderBalance := sender.Balance - amount
	newReceiverBalance := receiver.Balance + amount

	err = s.users.UpdateBalanceWithTx(trx, sender.Address, newSenderBalance)
	if err != nil {
		return models.Transaction{}, err
	}

	err = s.users.UpdateBalanceWithTx(trx, receiver.Address, newReceiverBalance)
	if err != nil {
		return models.Transaction{}, err
	}

	// create ledger entries
	err = s.ledgers.CreateWithTx(trx, txID, sender.Address, -amount, newSenderBalance)
	if err != nil {
		return models.Transaction{}, err
	}

	err = s.ledgers.CreateWithTx(trx, txID, receiver.Address, amount, newReceiverBalance)
	if err != nil {
		return models.Transaction{}, err
	}

	// commit transaction
	err = trx.Commit()
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.ID = txID
	tx.Status = "PENDING"

	return tx, nil
}
