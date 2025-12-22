package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type TransactionService interface {
	Send(fromAddress, toAddress, privateKey string, amount float64) (models.Transaction, error)
	GetTransactionByID(id int64) (models.Transaction, error)
	GenerateTransactionNonce(ctx context.Context, address string) string
	SendWithSignature(ctx context.Context, fromAddress, toAddress string, amount float64, nonce, signature string) (models.Transaction, error)
	Buy(ctx context.Context, address, signature, nonce string, amount float64) (models.Transaction, error)
}

type transactionService struct {
	users    repository.UserRepository
	txs      repository.TransactionRepository
	ledgers  repository.LedgerRepository
	memory   redis.MemoryAdapter
	txVerify VerifyTxService
}

func NewTransactionService(
	users repository.UserRepository,
	txs repository.TransactionRepository,
	ledgers repository.LedgerRepository,
	memory redis.MemoryAdapter,
	txVerify VerifyTxService,
) TransactionService {
	return &transactionService{
		users:    users,
		txs:      txs,
		ledgers:  ledgers,
		memory:   memory,
		txVerify: txVerify,
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
	isValid, err := utils.ValidatePrivateKeyMatchesAddress(privateKey, fromAddress)

	if err != nil && !isValid {
		return models.Transaction{}, fmt.Errorf("invalid private key for the given from address")
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

func (s *transactionService) GetTransactionByID(id int64) (models.Transaction, error) {
	return s.txs.GetTransactionByID(id)
}

func (s *transactionService) GenerateTransactionNonce(ctx context.Context, address string) string {
	addr := strings.ToLower(address)
	nonce := uuid.New().String()

	key := "tx_nonce:" + addr

	s.memory.Set(ctx, key, []byte(nonce), 5*time.Minute)

	return nonce
}

func (s *transactionService) SendWithSignature(ctx context.Context, fromAddress, toAddress string, amount float64, nonce, signature string) (models.Transaction, error) {
	// validate inputs amount
	if amount <= 0 {
		return models.Transaction{}, fmt.Errorf("amount must be greater than zero")
	}

	if fromAddress == toAddress {
		return models.Transaction{}, fmt.Errorf("cannot send to the same address")
	}

	// verify signature
	if err := s.txVerify.VerifyTransactionSignature(ctx, fromAddress, toAddress, amount, nonce, signature); err != nil {
		return models.Transaction{}, fmt.Errorf("signature verification failed: %w", err)
	}

	// get sender and receiver
	sender, err := s.users.GetByAddress(fromAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("sender not found")
	}

	_, err = s.users.GetByAddress(toAddress)

	if err != nil {
		return models.Transaction{}, fmt.Errorf("receiver not found")
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

	tx := models.Transaction{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Fee:         fee,
		Type:        "TRANSFER",
		Signature:   signature,
		Status:      "PENDING",
	}

	txID, err := s.txs.Create(tx)

	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to create tx: %w", err)
	}

	tx.ID = txID

	return tx, nil
}

func (s *transactionService) Buy(ctx context.Context, address, nonce, signature string, amount float64) (models.Transaction, error) {
	// validate inputs amount
	if amount <= 0 {
		return models.Transaction{}, fmt.Errorf("amount must be greater than zero")
	}

	//sistem sebagai penjual adalah address "MINER_ACCOUNT"
	buyerAddress := address
	sellerAddress := "MINER_ACCOUNT"

	// verify user exists
	_, err := s.users.GetByAddress(buyerAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("user not found")
	}

	// ambil akun miner
	_, err = s.users.GetByAddress(sellerAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("miner account not found")
	}

	// verify signature
	if err := s.txVerify.VerifyTransactionSignature(ctx, sellerAddress, buyerAddress, amount, nonce, signature); err != nil {
		return models.Transaction{}, fmt.Errorf("signature verification failed: %w", err)
	}

	// calculate transaction fee
	fee := utils.CalculateTransactionFee(amount)
	fee = utils.FormatFee(fee)

	// create transaction
	tx := models.Transaction{
		FromAddress: sellerAddress,
		ToAddress:   buyerAddress,
		Amount:      amount,
		Fee:         fee,
		Type:        "BUY",
		Signature:   signature,
		Status:      "PENDING",
	}

	txID, err := s.txs.Create(tx)

	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to create tx: %w", err)
	}

	tx.ID = txID

	// verify signature
	return tx, nil
}
