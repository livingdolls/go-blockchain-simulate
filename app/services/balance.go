package services

import (
	"errors"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

type BalanceService interface {
	GetBalance(address string) (models.User, error)
	GetWalletBalance(filter models.TransactionFilter) (models.WalletResponse, error)
}

type balanceService struct {
	users repository.UserRepository
	tx    repository.TransactionRepository
}

func NewBalanceService(users repository.UserRepository, tx repository.TransactionRepository) BalanceService {
	return &balanceService{
		users: users,
		tx:    tx,
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
		Ballance:     user.Balance,
		Address:      user.Address,
		Transactions: transaction,
	}

	return walletResponse, nil
}
