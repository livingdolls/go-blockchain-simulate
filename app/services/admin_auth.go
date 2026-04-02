package services

import (
	"context"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

type AdminAuthService interface {
	AuthenticateAdmin(ctx context.Context, username, password string) (*models.AdminWithUser, error)
}

type adminAuthService struct {
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
}

func NewAdminAuthService(userRepo repository.UserRepository, adminRepo repository.AdminRepository) AdminAuthService {
	return &adminAuthService{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

func (s *adminAuthService) AuthenticateAdmin(ctx context.Context, username, password string) (*models.AdminWithUser, error) {
	admin, err := s.adminRepo.GetAdminByUsernameWithPassword(ctx, username)
	if err != nil {
		return nil, err
	}

	if !security.CheckPasswordHash(admin.PasswordHash, password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	if admin.Status != "active" {
		return nil, fmt.Errorf("admin account is not active")
	}

	// map AdminWithPassword to AdminWithUser
	adminWithUser := &models.AdminWithUser{
		ID:          admin.ID,
		UserID:      admin.UserID,
		Username:    admin.Username,
		Address:     admin.Address,
		Role:        admin.Role,
		Permissions: admin.Permissions,
		Status:      admin.Status,
		LastLoginAt: admin.LastLoginAt,
		CreatedAt:   admin.CreatedAt,
	}

	return adminWithUser, nil
}
