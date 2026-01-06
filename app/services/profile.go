package services

import (
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

type ProfileService interface {
	Me(address string) (models.User, error)
}

type profileService struct {
	repo repository.UserRepository
}

func NewProfileService(repo repository.UserRepository) ProfileService {
	return &profileService{
		repo: repo,
	}
}

func (s *profileService) Me(address string) (models.User, error) {
	return s.repo.GetByAddress(address)
}

func (s *profileService) MeWithBalance(address string) (dto.DTOUserWithBalance, error) {
	userWithBalance, err := s.repo.GetByAddressWithBalance(address)
	if err != nil {
		return dto.DTOUserWithBalance{}, err
	}

	dtoUser := dto.DTOUserWithBalance{
		Name:       userWithBalance.Name,
		Address:    userWithBalance.Address,
		YTEBalance: userWithBalance.YTEBalance,
		USDBalance: userWithBalance.USDBalance,
	}

	return dtoUser, nil
}
