package services

import (
	"errors"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"go.uber.org/zap"
)

type BalanceService interface {
	GetBalance(address string) (models.User, error)
	GetUserWithUSDBalance(address string) (dto.DTOUserWithBalance, error)
	GetWalletBalance(filter models.TransactionFilter) (models.WalletResponse, error)
	TopUpUSDBalance(address string, amount float64, referenceID, description string) (dto.TopUpResultDTO, error)
}

type balanceService struct {
	users        repository.UserRepository
	tx           repository.TransactionRepository
	userBalances repository.UserBalanceRepository
	publisherWs  *publisher.PublisherWS
}

func NewBalanceService(users repository.UserRepository, tx repository.TransactionRepository, userBalances repository.UserBalanceRepository, publisherWs *publisher.PublisherWS) BalanceService {
	return &balanceService{
		users:        users,
		tx:           tx,
		userBalances: userBalances,
		publisherWs:  publisherWs,
	}
}

func (s *balanceService) GetBalance(address string) (models.User, error) {
	user, err := s.users.GetByAddress(address)
	if err != nil {
		return models.User{}, errors.New("address not found")
	}
	return user, nil
}

func (s *balanceService) GetWalletBalance(filter models.TransactionFilter) (models.WalletResponse, error) {
	user, err := s.users.GetByAddress(filter.Address)
	if err != nil {
		return models.WalletResponse{}, errors.New("address not found")
	}

	transaction, err := s.tx.GetTransactionByAddress(filter)

	if err != nil {
		return models.WalletResponse{}, errors.New("could not retrieve transactions")
	}

	walletResponse := models.WalletResponse{
		Address:      user.Address,
		Transactions: transaction,
	}

	return walletResponse, nil
}

func (s *balanceService) TopUpUSDBalance(address string, amount float64, referenceID, description string) (dto.TopUpResultDTO, error) {
	if address == "" {
		return dto.TopUpResultDTO{}, entity.ErrAddressNotFound
	}

	if amount <= 0 {
		return dto.TopUpResultDTO{}, entity.ErrAmountMustBePositive
	}

	// pastikan user ada
	if _, err := s.users.GetByAddress(address); err != nil {
		return dto.TopUpResultDTO{}, entity.ErrAddressNotFound
	}

	tx, err := s.userBalances.BeginTx()
	if err != nil {
		return dto.TopUpResultDTO{}, fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := s.userBalances.UpsertEmptyIfNotExistsWithTx(tx, address); err != nil {
		return dto.TopUpResultDTO{}, fmt.Errorf("upsert empty user balance: %v", err)
	}

	ub, err := s.userBalances.GetForUpdateWithTx(tx, address)

	if err != nil {
		return dto.TopUpResultDTO{}, fmt.Errorf("get user balance for update: %v", err)
	}

	balanceBefore := ub.USDBalance
	afterBalance := balanceBefore + amount
	totalDeposited := ub.TotalDeposited + amount

	if err := s.userBalances.UpdateBalanceWithTx(tx, address, afterBalance, totalDeposited); err != nil {
		return dto.TopUpResultDTO{}, fmt.Errorf("update user balance: %v", err)
	}

	var refPtr *string
	var descPtr *string

	if referenceID != "" {
		refPtr = &referenceID
	}

	if description != "" {
		descPtr = &description
	}

	history := models.BalanceHistory{
		UserAddress:   address,
		ChangeType:    "DEPOSIT",
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  afterBalance,
		LockedBefore:  ub.LockedBalance,
		LockedAfter:   ub.LockedBalance,
		ReferenceID:   refPtr,
		Description:   descPtr,
	}

	if err := s.userBalances.InsertHistoryWithTx(tx, history); err != nil {
		return dto.TopUpResultDTO{}, fmt.Errorf("insert balance history: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.TopUpResultDTO{}, fmt.Errorf("commit tx: %v", err)
	}

	result := dto.TopUpResultDTO{
		Address:       address,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  afterBalance,
		ReferenceID:   refPtr,
		Description:   descPtr,
	}

	// Notify via WebSocket
	// prepare data
	data, err := s.users.GetByAddressWithBalance(address)
	if err != nil {
		// don't fail the top-up just because notification fails
		logger.LogError("failed to get user data for notification", err)
	} else {
		logger.LogInfo("Publishing balance update for address", zap.String("address", address), zap.Any("data", data))
		dtoResult := dto.DTOUserWithBalance{
			Name:       data.Name,
			Address:    data.Address,
			YTEBalance: data.YTEBalance,
			USDBalance: data.USDBalance,
		}

		s.publisherWs.PublishToAddress(address, entity.EventBalanceUpdate, dtoResult)
	}

	return result, nil
}

func (s *balanceService) GetUserWithUSDBalance(address string) (dto.DTOUserWithBalance, error) {
	if address == "" {
		return dto.DTOUserWithBalance{}, entity.ErrAddressNotFound
	}

	user, err := s.users.GetByAddressWithBalance(address)

	if err != nil {
		return dto.DTOUserWithBalance{}, entity.ErrAddressNotFound
	}

	result := dto.DTOUserWithBalance{
		Name:       user.Name,
		Address:    user.Address,
		USDBalance: user.USDBalance,
		YTEBalance: user.YTEBalance,
	}

	return result, nil
}
