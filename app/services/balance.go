package services

import (
	"errors"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

type BalanceService interface {
	GetBalance(address string) (models.User, error)
}

type balanceService struct {
	users repository.UserRepository
}

func NewBalanceService(users repository.UserRepository) BalanceService {
	return &balanceService{
		users: users,
	}
}

func (s *balanceService) GetBalance(address string) (models.User, error) {
	user, err := s.users.GetByAddress(address)
	if err != nil {
		return models.User{}, errors.New("address not found")
	}
	return user, nil
}
