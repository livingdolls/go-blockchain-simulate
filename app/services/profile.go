package services

import (
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
