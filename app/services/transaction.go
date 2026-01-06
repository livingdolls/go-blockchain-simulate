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
	GetTransactionByID(id int64) (models.Transaction, error)
	GenerateTransactionNonce(ctx context.Context, address string) string
	SendWithSignature(ctx context.Context, fromAddress, toAddress string, amount float64, nonce, signature string) (models.Transaction, error)
	Buy(ctx context.Context, address, signature, nonce string, amount float64) (models.Transaction, error)
	Sell(ctx context.Context, address, signature, nonce string, amount float64) (models.Transaction, error)
}

type transactionService struct {
	users    repository.UserRepository
	wallets  repository.UserWalletRepository
	txs      repository.TransactionRepository
	ledgers  repository.LedgerRepository
	memory   redis.MemoryAdapter
	txVerify VerifyTxService
}

func NewTransactionService(
	users repository.UserRepository,
	wallets repository.UserWalletRepository,
	txs repository.TransactionRepository,
	ledgers repository.LedgerRepository,
	memory redis.MemoryAdapter,
	txVerify VerifyTxService,
) TransactionService {
	return &transactionService{
		users:    users,
		wallets:  wallets,
		txs:      txs,
		ledgers:  ledgers,
		memory:   memory,
		txVerify: txVerify,
	}
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
	senderWallet, err := s.ensureWallet(fromAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("sender wallet not found for address %s", fromAddress)
	}

	// ensure user and wallet exists for receiver
	_, receiverWallet, err := s.users.GetUserWithWallet(toAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("receiver not found")
	}

	if receiverWallet.UserAddress == "" {
		_, err = s.ensureWallet(toAddress)
		if err != nil {
			return models.Transaction{}, fmt.Errorf("failed to create receiver wallet: %w", err)
		}
		// Refetch receiver wallet after creation to validate
		_, receiverWallet, err = s.users.GetUserWithWallet(toAddress)
		if err != nil {
			return models.Transaction{}, fmt.Errorf("failed to retrieve receiver wallet: %w", err)
		}
	}

	// calculate transaction fee
	fee := utils.CalculateTransactionFee(amount)
	fee = utils.FormatFee(fee)

	// calculate total deduction
	totalRequired := amount + fee

	// check if sender has enough balance
	if senderWallet.YTEBalance < totalRequired {
		return models.Transaction{}, fmt.Errorf("insufficient balance: required %.5f, available %.5f", totalRequired, senderWallet.YTEBalance)
	}

	// check pending transactions from sender to prevent double spending
	pendingAmount, err := s.txs.GetPendingTransactionsByAddress(fromAddress)
	if err == nil && pendingAmount > 0 {
		availableBalance := senderWallet.YTEBalance - pendingAmount
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

	// ensure buyer wallet exists
	_, buyerWallet, err := s.users.GetUserWithWallet(buyerAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("buyer wallet not found for address %s", buyerAddress)
	}

	// create wallet if not exists
	if buyerWallet.UserAddress == "" {
		_, err = s.ensureWallet(buyerAddress)
		if err != nil {
			return models.Transaction{}, fmt.Errorf("buyer wallet not found for address %s", buyerAddress)
		}

		_, buyerWallet, err = s.users.GetUserWithWallet(buyerAddress)
		if err != nil {
			return models.Transaction{}, fmt.Errorf("buyer wallet not found for address %s", buyerAddress)
		}
	}

	// check miner account
	_, err = s.ensureWallet(sellerAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("seller wallet not found for address %s", sellerAddress)
	}

	// verify signature
	if err := s.txVerify.VerifyBuySellSignature(ctx, buyerAddress, amount, nonce, signature, BuyTransaction); err != nil {
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
	tx.FromAddress = "SYSTEM_SELLER"

	return tx, nil
}

func (s *transactionService) Sell(ctx context.Context, address, nonce, signature string, amount float64) (models.Transaction, error) {
	// validate inputs amount
	if amount <= 0 {
		return models.Transaction{}, fmt.Errorf("amount must be greater than zero")
	}

	//sistem sebagai pembeli adalah address "MINER_ACCOUNT"
	sellerAddress := address
	buyerAddress := "MINER_ACCOUNT"

	// verify user exists
	_, sellerWallet, err := s.users.GetUserWithWallet(sellerAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("user not found")
	}

	// create wallet if not exists
	if sellerWallet.UserAddress == "" {
		_, err = s.ensureWallet(sellerAddress)
		if err != nil {
			return models.Transaction{}, fmt.Errorf("seller wallet not found for address %s", sellerAddress)
		}

		_, sellerWallet, err = s.users.GetUserWithWallet(sellerAddress)
		if err != nil {
			return models.Transaction{}, fmt.Errorf("seller wallet not found for address %s", sellerAddress)
		}
	}

	// check miner account
	_, err = s.ensureWallet(buyerAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("buyer wallet not found for address %s", buyerAddress)
	}

	// verify signature
	if err := s.txVerify.VerifyBuySellSignature(ctx, sellerAddress, amount, nonce, signature, SellTransaction); err != nil {
		return models.Transaction{}, fmt.Errorf("signature verification failed: %w", err)
	}

	// check if seller has enough balance
	if sellerWallet.YTEBalance < amount {
		return models.Transaction{}, fmt.Errorf("insufficient balance: required %.5f, available %.5f", amount, sellerWallet.YTEBalance)
	}

	// check pending transactions from seller to prevent double spending
	pendingAmount, err := s.txs.GetPendingTransactionsByAddress(sellerAddress)
	if err == nil && pendingAmount > 0 {
		availableBalance := sellerWallet.YTEBalance - pendingAmount
		if availableBalance < amount {
			return models.Transaction{}, fmt.Errorf("insufficient balance considering pending transactions: required %.5f, available %.5f", amount, availableBalance)
		}
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
		Type:        "SELL",
		Signature:   signature,
		Status:      "PENDING",
	}

	txID, err := s.txs.Create(tx)

	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to create tx: %w", err)
	}

	tx.ID = txID
	tx.ToAddress = "SYSTEM_BUYER"
	return tx, nil
}

func (s *transactionService) ensureWallet(address string) (models.UserWallet, error) {
	wallet, err := s.wallets.GetByAddress(address)
	if err == nil {
		return wallet, nil
	}

	// Wallet not found, attempt to create it
	tx, err := s.wallets.BeginTx()
	if err != nil {
		return models.UserWallet{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.wallets.UpsertEmptyIfNotExistsWithTx(tx, address); err != nil {
		return models.UserWallet{}, fmt.Errorf("upsert empty wallet: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return models.UserWallet{}, fmt.Errorf("commit tx: %w", err)
	}

	// Retrieve wallet after creation
	wallet, err = s.wallets.GetByAddress(address)
	if err != nil {
		return models.UserWallet{}, fmt.Errorf("get wallet after upsert: %w", err)
	}

	return wallet, nil
}
