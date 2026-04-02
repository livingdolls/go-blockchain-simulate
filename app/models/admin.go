package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

func nullStringValue(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func nullTimeValue(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	v := value.Time.Format(time.RFC3339)
	return &v
}

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

func (a Admin) MarshalJSON() ([]byte, error) {
	type adminJSON struct {
		ID          int     `json:"id"`
		UserID      int     `json:"user_id"`
		Role        string  `json:"role"`
		Permissions *string `json:"permissions"`
		Status      string  `json:"status"`
		LastLoginAt *string `json:"last_login_at"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
	}

	return json.Marshal(adminJSON{
		ID:          a.ID,
		UserID:      a.UserID,
		Role:        a.Role,
		Permissions: nullStringValue(a.Permissions),
		Status:      a.Status,
		LastLoginAt: nullTimeValue(a.LastLoginAt),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	})
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

func (l AdminActivityLog) MarshalJSON() ([]byte, error) {
	type adminActivityLogJSON struct {
		ID             int64   `json:"id"`
		AdminID        int     `json:"admin_id"`
		Action         string  `json:"action"`
		TargetEntity   *string `json:"target_entity"`
		TargetID       *string `json:"target_id"`
		TargetName     *string `json:"target_name"`
		OldValues      *string `json:"old_values"`
		NewValues      *string `json:"new_values"`
		ChangesSummary *string `json:"changes_summary"`
		IPAddress      *string `json:"ip_address"`
		UserAgent      *string `json:"user_agent"`
		Status         string  `json:"status"`
		ErrorMessage   *string `json:"error_message"`
		CreatedAt      string  `json:"created_at"`
	}

	return json.Marshal(adminActivityLogJSON{
		ID:             l.ID,
		AdminID:        l.AdminID,
		Action:         l.Action,
		TargetEntity:   nullStringValue(l.TargetEntity),
		TargetID:       nullStringValue(l.TargetID),
		TargetName:     nullStringValue(l.TargetName),
		OldValues:      nullStringValue(l.OldValues),
		NewValues:      nullStringValue(l.NewValues),
		ChangesSummary: nullStringValue(l.ChangesSummary),
		IPAddress:      nullStringValue(l.IPAddress),
		UserAgent:      nullStringValue(l.UserAgent),
		Status:         l.Status,
		ErrorMessage:   nullStringValue(l.ErrorMessage),
		CreatedAt:      l.CreatedAt,
	})
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

func (a AdminWithUser) MarshalJSON() ([]byte, error) {
	type adminWithUserJSON struct {
		ID          int     `json:"id"`
		UserID      int     `json:"user_id"`
		Username    string  `json:"username"`
		Address     string  `json:"address"`
		Role        string  `json:"role"`
		Permissions *string `json:"permissions"`
		Status      string  `json:"status"`
		LastLoginAt *string `json:"last_login_at"`
		CreatedAt   string  `json:"created_at"`
	}

	return json.Marshal(adminWithUserJSON{
		ID:          a.ID,
		UserID:      a.UserID,
		Username:    a.Username,
		Address:     a.Address,
		Role:        a.Role,
		Permissions: nullStringValue(a.Permissions),
		Status:      a.Status,
		LastLoginAt: nullTimeValue(a.LastLoginAt),
		CreatedAt:   a.CreatedAt,
	})
}

type AdminWithPassword struct {
	AdminWithUser
	PasswordHash string `db:"password_hash" json:"-"`
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
