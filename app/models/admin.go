package models

import "database/sql"

// Admin represents an admin user
type Admin struct {
	ID          int            `db:"id" json:"id"`
	UserID      int            `db:"user_id" json:"user_id"`
	Role        string         `db:"role" json:"role"` // admin, moderator, support
	Permissions sql.NullString `db:"permissions" json:"permissions"`
	Status      string         `db:"status" json:"status"` // active, inactive, suspended
	LastLoginAt sql.NullTime   `db:"last_login_at" json:"last_login_at"`
	CreatedAt   string         `db:"created_at" json:"created_at"`
	UpdatedAt   string         `db:"updated_at" json:"updated_at"`
}

// AdminActivityLog represents an action performed by admin
type AdminActivityLog struct {
	ID             int64          `db:"id" json:"id"`
	AdminID        int            `db:"admin_id" json:"admin_id"`
	Action         string         `db:"action" json:"action"` // create, update, delete, approve, reject
	TargetEntity   sql.NullString `db:"target_entity" json:"target_entity"`
	TargetID       sql.NullString `db:"target_id" json:"target_id"`
	TargetName     sql.NullString `db:"target_name" json:"target_name"`
	OldValues      sql.NullString `db:"old_values" json:"old_values"`
	NewValues      sql.NullString `db:"new_values" json:"new_values"`
	ChangesSummary sql.NullString `db:"changes_summary" json:"changes_summary"`
	IPAddress      sql.NullString `db:"ip_address" json:"ip_address"`
	UserAgent      sql.NullString `db:"user_agent" json:"user_agent"`
	Status         string         `db:"status" json:"status"` // success, failed, pending
	ErrorMessage   sql.NullString `db:"error_message" json:"error_message"`
	CreatedAt      string         `db:"created_at" json:"created_at"`
}

// AdminWithUser combines Admin and User info
type AdminWithUser struct {
	ID          int            `db:"id" json:"id"`
	UserID      int            `db:"user_id" json:"user_id"`
	Username    string         `db:"name" json:"username"`
	Address     string         `db:"address" json:"address"`
	Role        string         `db:"role" json:"role"`
	Permissions sql.NullString `db:"permissions" json:"permissions"`
	Status      string         `db:"status" json:"status"`
	LastLoginAt sql.NullTime   `db:"last_login_at" json:"last_login_at"`
	CreatedAt   string         `db:"created_at" json:"created_at"`
}

// AdminDashboardStats untuk dashboard admin
type AdminDashboardStats struct {
	TotalUsers          int64   `json:"total_users"`
	TotalTransactions   int64   `json:"total_transactions"`
	TotalBlocks         int64   `json:"total_blocks"`
	TotalAdmins         int     `json:"total_admins"`
	ActiveUsers         int64   `json:"active_users"`
	SuspendedAdmins     int     `json:"suspended_admins"`
	RecentActivityCount int64   `json:"recent_activity_count"`
	TotalVolume         float64 `json:"total_volume"`
}

// PermissionSet untuk cek permission
var PermissionMap = map[string][]string{
	"admin": {
		"*", // wildcard untuk all permissions
	},
	"moderator": {
		"read_users",
		"read_transactions",
		"read_blocks",
		"moderate_users",
		"moderate_transactions",
		"view_activity_logs",
	},
	"support": {
		"read_users",
		"read_transactions",
		"user_support",
		"view_activity_logs",
	},
}
