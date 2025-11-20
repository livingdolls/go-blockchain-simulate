package services

import (
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type RegisterService interface {
	Registr(name string) (models.User, error)
}

type registerService struct {
	repo repository.UserRepository
}

func NewRegisterService(repo repository.UserRepository) RegisterService {
	return &registerService{repo: repo}
}

// Registr implements RegisterService.
func (r *registerService) Registr(name string) (models.User, error) {
	privateKey, publicKey := utils.GenerateFakeKey()
	address := utils.GenerateAddressFromPublicKey(publicKey)

	user := models.User{
		Name:       name,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		Balance:    1000000,
	}

	err := r.repo.Create(user)

	return user, err
}
