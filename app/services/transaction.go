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
	// validate inputs amount
	if amount <= 0 {
		return models.Transaction{}, fmt.Errorf("amount must be greater than zero")
	}

	if fromAddress == toAddress {
		return models.Transaction{}, fmt.Errorf("cannot send to the same address")
	}

	// first, get sender and receiver
	sender, err := s.users.GetByAddress(fromAddress)
	if err != nil {
		return models.Transaction{}, err
	}

	receiver, err := s.users.GetByAddress(toAddress)

	if err != nil {
		return models.Transaction{}, err
	}

	// calculate transaction fee
	fee := utils.CalculateTransactionFee(amount)
	fee = utils.FormatFee(fee)

	// calculate total deduction
	totalRequired := amount + fee

	// check if sender has enough balance
	if sender.Balance < totalRequired {
		return models.Transaction{}, fmt.Errorf("insufficient balance: required %.5f, available %.5f", totalRequired, sender.Balance)
	}

	// check pending transactions from sender to prevent double spending
	pendingAmount, err := s.txs.GetPendingTransactionsByAddress(fromAddress)
	if err == nil && pendingAmount > 0 {
		availableBalance := sender.Balance - pendingAmount
		if availableBalance < totalRequired {
			return models.Transaction{}, fmt.Errorf("insufficient balance considering pending transactions: required %.5f, available %.5f", totalRequired, availableBalance)
		}
	}

	// validate private key
	if sender.PrivateKey != privateKey {
		return models.Transaction{}, fmt.Errorf("invalid private key for address %s", fromAddress)
	}

	// create digital signature
	signature := utils.SignFake(privateKey, receiver.Address, amount)
	// create transaction
	tx := models.Transaction{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Fee:         fee,
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
