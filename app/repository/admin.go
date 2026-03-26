package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

// AdminRepository handles admin-related database operations
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates new admin repository
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetAdminByUserID fetches admin by user ID
func (r *AdminRepository) GetAdminByUserID(userID int) (*models.AdminWithUser, error) {
	query := `
		SELECT 
			a.id, a.user_id, u.name, u.address, a.role, a.permissions, 
			a.status, a.last_login_at, a.created_at
		FROM admins a
		JOIN users u ON a.user_id = u.id
		WHERE a.user_id = ? AND a.status = 'active'
	`

	admin := &models.AdminWithUser{}
	err := r.db.QueryRow(query, userID).Scan(
		&admin.ID, &admin.UserID, &admin.Username, &admin.Address,
		&admin.Role, &admin.Permissions, &admin.Status, &admin.LastLoginAt,
		&admin.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, err
	}

	return admin, nil
}

// GetAllAdmins fetches all admin users with pagination
func (r *AdminRepository) GetAllAdmins(limit, offset int) ([]*models.AdminWithUser, error) {
	query := `
		SELECT 
			a.id, a.user_id, u.name, u.address, a.role, a.permissions, 
			a.status, a.last_login_at, a.created_at
		FROM admins a
		JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []*models.AdminWithUser
	for rows.Next() {
		admin := &models.AdminWithUser{}
		err := rows.Scan(
			&admin.ID, &admin.UserID, &admin.Username, &admin.Address,
			&admin.Role, &admin.Permissions, &admin.Status, &admin.LastLoginAt,
			&admin.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}

	return admins, rows.Err()
}

// CreateAdmin creates new admin user
func (r *AdminRepository) CreateAdmin(userID int, role string, permissions []string) error {
	permJSON, _ := json.Marshal(permissions)

	query := `
		INSERT INTO admins (user_id, role, permissions, status)
		VALUES (?, ?, ?, 'active')
	`

	_, err := r.db.Exec(query, userID, role, string(permJSON))
	return err
}

// UpdateAdminRole updates admin role and permissions
func (r *AdminRepository) UpdateAdminRole(adminID int, role string, permissions []string) error {
	permJSON, _ := json.Marshal(permissions)

	query := `
		UPDATE admins
		SET role = ?, permissions = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.Exec(query, role, string(permJSON), adminID)
	return err
}

// UpdateAdminStatus updates admin status
func (r *AdminRepository) UpdateAdminStatus(adminID int, status string) error {
	query := `
		UPDATE admins
		SET status = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.Exec(query, status, adminID)
	return err
}

// UpdateLastLogin updates last login timestamp
func (r *AdminRepository) UpdateLastLogin(adminID int) error {
	query := `UPDATE admins SET last_login_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, adminID)
	return err
}

// LogActivity logs admin action to activity log
func (r *AdminRepository) LogActivity(log *models.AdminActivityLog) error {
	query := `
		INSERT INTO admin_activity_logs (
			admin_id, action, target_entity, target_id, target_name,
			old_values, new_values, changes_summary, ip_address, user_agent, status, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		log.AdminID, log.Action, log.TargetEntity, log.TargetID, log.TargetName,
		log.OldValues, log.NewValues, log.ChangesSummary, log.IPAddress, log.UserAgent,
		log.Status, log.ErrorMessage,
	)

	return err
}

// GetActivityLogs fetches admin activity logs with filters
func (r *AdminRepository) GetActivityLogs(adminID int, action string, limit, offset int) ([]*models.AdminActivityLog, error) {
	query := `
		SELECT 
			id, admin_id, action, target_entity, target_id, target_name,
			old_values, new_values, changes_summary, ip_address, user_agent, status, error_message, created_at
		FROM admin_activity_logs
		WHERE admin_id = ?
	`

	if action != "" {
		query += ` AND action = '` + action + `'`
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, adminID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AdminActivityLog
	for rows.Next() {
		log := &models.AdminActivityLog{}
		err := rows.Scan(
			&log.ID, &log.AdminID, &log.Action, &log.TargetEntity, &log.TargetID, &log.TargetName,
			&log.OldValues, &log.NewValues, &log.ChangesSummary, &log.IPAddress, &log.UserAgent,
			&log.Status, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetRecentActivityLogs fetches recent activity logs for dashboard
func (r *AdminRepository) GetRecentActivityLogs(days int, limit int) ([]*models.AdminActivityLog, error) {
	query := `
		SELECT 
			id, admin_id, action, target_entity, target_id, target_name,
			old_values, new_values, changes_summary, ip_address, user_agent, status, error_message, created_at
		FROM admin_activity_logs
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, days, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AdminActivityLog
	for rows.Next() {
		log := &models.AdminActivityLog{}
		err := rows.Scan(
			&log.ID, &log.AdminID, &log.Action, &log.TargetEntity, &log.TargetID, &log.TargetName,
			&log.OldValues, &log.NewValues, &log.ChangesSummary, &log.IPAddress, &log.UserAgent,
			&log.Status, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetDashboardStats fetches dashboard statistics
func (r *AdminRepository) GetDashboardStats() (*models.AdminDashboardStats, error) {
	stats := &models.AdminDashboardStats{}

	// Total users
	_ = r.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'user'").Scan(&stats.TotalUsers)

	// Total transactions
	_ = r.db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&stats.TotalTransactions)

	// Total blocks
	_ = r.db.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&stats.TotalBlocks)

	// Total admins
	_ = r.db.QueryRow("SELECT COUNT(*) FROM admins WHERE status = 'active'").Scan(&stats.TotalAdmins)

	// Active users (last 7 days)
	_ = r.db.QueryRow(`
		SELECT COUNT(DISTINCT user_address) 
		FROM transactions 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
	`).Scan(&stats.ActiveUsers)

	// Suspended admins
	_ = r.db.QueryRow("SELECT COUNT(*) FROM admins WHERE status = 'suspended'").Scan(&stats.SuspendedAdmins)

	// Recent activity count (last 24 hours)
	_ = r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM admin_activity_logs 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
	`).Scan(&stats.RecentActivityCount)

	// Total volume from transactions
	_ = r.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions").Scan(&stats.TotalVolume)

	return stats, nil
}

// CountAdmins returns total admin count
func (r *AdminRepository) CountAdmins() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM admins WHERE status != 'inactive'").Scan(&count)
	return count, err
}

// DeleteAdmin soft deletes admin by marking as inactive
func (r *AdminRepository) DeleteAdmin(adminID int) error {
	query := `UPDATE admins SET status = 'inactive', updated_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, adminID)
	return err
}
