package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

// AdminService defines admin operations
type AdminService interface {
	// Dashboard
	GetDashboardStats(ctx context.Context, admin *models.AdminWithUser) (*models.AdminDashboardStats, error)
	GetRecentActivityLogs(ctx context.Context, admin *models.AdminWithUser, days, limit int) ([]*models.AdminActivityLog, error)

	// Admin management
	GetAllAdmins(ctx context.Context, admin *models.AdminWithUser, limit, offset int) ([]*models.AdminWithUser, error)
	CreateAdmin(ctx context.Context, admin *models.AdminWithUser, userID int, role string, permissions []string) error
	UpdateAdminRole(ctx context.Context, admin *models.AdminWithUser, targetAdminID int, role string, permissions []string) error
	UpdateAdminStatus(ctx context.Context, admin *models.AdminWithUser, targetAdminID int, status string) error
	DeleteAdmin(ctx context.Context, admin *models.AdminWithUser, targetAdminID int) error

	// Activity logs
	LogActivity(ctx context.Context, log *models.AdminActivityLog) error
	GetActivityLogs(ctx context.Context, admin *models.AdminWithUser, targetAdminID int, action string, limit, offset int) ([]*models.AdminActivityLog, error)
}

type adminService struct {
	repo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

// GetDashboardStats retrieves dashboard statistics with permission check
func (s *adminService) GetDashboardStats(ctx context.Context, admin *models.AdminWithUser) (*models.AdminDashboardStats, error) {
	if !s.hasPermission(admin, "read_dashboard") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	return s.repo.GetDashboardStats(ctx)
}

// GetRecentActivityLogs retrieves recent admin activity logs
func (s *adminService) GetRecentActivityLogs(ctx context.Context, admin *models.AdminWithUser, days, limit int) ([]*models.AdminActivityLog, error) {
	if !s.hasPermission(admin, "view_activity_logs") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	if limit > 1000 {
		limit = 1000
	}
	return s.repo.GetRecentActivityLogs(ctx, days, limit)
}

// GetAllAdmins retrieves all admin users
func (s *adminService) GetAllAdmins(ctx context.Context, admin *models.AdminWithUser, limit, offset int) ([]*models.AdminWithUser, error) {
	if !s.hasPermission(admin, "manage_admins") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	if limit > 100 {
		limit = 100
	}
	return s.repo.GetAllAdmins(ctx, limit, offset)
}

// CreateAdmin creates new admin user
func (s *adminService) CreateAdmin(ctx context.Context, admin *models.AdminWithUser, userID int, role string, permissions []string) error {
	if !s.hasPermission(admin, "manage_admins") {
		return fmt.Errorf("insufficient permissions")
	}

	if !s.isValidRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}

	log := &models.AdminActivityLog{
		AdminID:      admin.ID,
		Action:       "create_admin",
		TargetEntity: sql.NullString{String: "admins", Valid: true},
		TargetID:     sql.NullString{String: fmt.Sprintf("%d", userID), Valid: true},
		Status:       "pending",
	}

	err := s.repo.CreateAdmin(ctx, userID, role, permissions)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = sql.NullString{String: err.Error(), Valid: true}
		s.repo.LogActivity(ctx, log)
		return err
	}

	log.Status = "success"
	s.repo.LogActivity(ctx, log)
	return nil
}

// UpdateAdminRole updates admin role and permissions
func (s *adminService) UpdateAdminRole(ctx context.Context, admin *models.AdminWithUser, targetAdminID int, role string, permissions []string) error {
	if !s.hasPermission(admin, "manage_admins") {
		return fmt.Errorf("insufficient permissions")
	}

	if !s.isValidRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Prevent self-demotion
	if admin.ID == targetAdminID && role != "admin" {
		return fmt.Errorf("cannot demote yourself")
	}

	log := &models.AdminActivityLog{
		AdminID:        admin.ID,
		Action:         "update_admin_role",
		TargetEntity:   sql.NullString{String: "admins", Valid: true},
		TargetID:       sql.NullString{String: fmt.Sprintf("%d", targetAdminID), Valid: true},
		ChangesSummary: sql.NullString{String: fmt.Sprintf("role changed to %s", role), Valid: true},
		Status:         "pending",
	}

	err := s.repo.UpdateAdminRole(ctx, targetAdminID, role, permissions)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = sql.NullString{String: err.Error(), Valid: true}
		s.repo.LogActivity(ctx, log)
		return err
	}

	log.Status = "success"
	s.repo.LogActivity(ctx, log)
	return nil
}

// UpdateAdminStatus updates admin status
func (s *adminService) UpdateAdminStatus(ctx context.Context, admin *models.AdminWithUser, targetAdminID int, status string) error {
	if !s.hasPermission(admin, "manage_admins") {
		return fmt.Errorf("insufficient permissions")
	}

	validStatuses := []string{"active", "inactive", "suspended"}
	if !s.contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	log := &models.AdminActivityLog{
		AdminID:        admin.ID,
		Action:         "update_admin_status",
		TargetEntity:   sql.NullString{String: "admins", Valid: true},
		TargetID:       sql.NullString{String: fmt.Sprintf("%d", targetAdminID), Valid: true},
		ChangesSummary: sql.NullString{String: fmt.Sprintf("status changed to %s", status), Valid: true},
		Status:         "pending",
	}

	err := s.repo.UpdateAdminStatus(ctx, targetAdminID, status)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = sql.NullString{String: err.Error(), Valid: true}
		s.repo.LogActivity(ctx, log)
		return err
	}

	log.Status = "success"
	s.repo.LogActivity(ctx, log)
	return nil
}

// DeleteAdmin deletes admin user (soft delete)
func (s *adminService) DeleteAdmin(ctx context.Context, admin *models.AdminWithUser, targetAdminID int) error {
	if !s.hasPermission(admin, "manage_admins") {
		return fmt.Errorf("insufficient permissions")
	}

	if admin.ID == targetAdminID {
		return fmt.Errorf("cannot delete yourself")
	}

	log := &models.AdminActivityLog{
		AdminID:      admin.ID,
		Action:       "delete_admin",
		TargetEntity: sql.NullString{String: "admins", Valid: true},
		TargetID:     sql.NullString{String: fmt.Sprintf("%d", targetAdminID), Valid: true},
		Status:       "pending",
	}

	err := s.repo.DeleteAdmin(ctx, targetAdminID)
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = sql.NullString{String: err.Error(), Valid: true}
		s.repo.LogActivity(ctx, log)
		return err
	}

	log.Status = "success"
	s.repo.LogActivity(ctx, log)
	return nil
}

// LogActivity logs admin action
func (s *adminService) LogActivity(ctx context.Context, log *models.AdminActivityLog) error {
	return s.repo.LogActivity(ctx, log)
}

// GetActivityLogs retrieves activity logs for admin
func (s *adminService) GetActivityLogs(ctx context.Context, admin *models.AdminWithUser, targetAdminID int, action string, limit, offset int) ([]*models.AdminActivityLog, error) {
	if !s.hasPermission(admin, "view_activity_logs") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Users can only see their own logs unless they are admin
	if admin.Role != "admin" && admin.ID != targetAdminID {
		return nil, fmt.Errorf("cannot view other admin's logs")
	}

	if limit > 500 {
		limit = 500
	}

	return s.repo.GetActivityLogs(ctx, targetAdminID, action, limit, offset)
}

// Helper functions

// hasPermission checks if admin has required permission
func (s *adminService) hasPermission(admin *models.AdminWithUser, permission string) bool {
	if admin.Role == "admin" {
		return true
	}

	rolePermissions := map[string][]string{
		"admin": {"*"},
		"moderator": {
			"read_dashboard",
			"view_activity_logs",
			"read_users",
			"read_transactions",
		},
		"support": {
			"read_users",
			"view_activity_logs",
		},
	}

	permissions, ok := rolePermissions[admin.Role]
	if !ok {
		return false
	}

	for _, perm := range permissions {
		if perm == "*" || perm == permission {
			return true
		}
	}
	return false
}

// isValidRole checks if role is valid
func (s *adminService) isValidRole(role string) bool {
	validRoles := []string{"admin", "moderator", "support"}
	return s.contains(validRoles, role)
}

// contains checks if slice contains string
func (s *adminService) contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
