-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    role VARCHAR(50) NOT NULL DEFAULT 'admin' COMMENT 'admin, moderator, support',
    permissions JSON COMMENT 'JSON array of permission strings',
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_role (role),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Add role column to users table if not exists
ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user' NOT NULL AFTER address;

-- Create admin_activity_logs table for audit trail
CREATE TABLE IF NOT EXISTS admin_activity_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    admin_id INT NOT NULL,
    action VARCHAR(100) NOT NULL COMMENT 'create, update, delete, approve, reject',
    target_entity VARCHAR(50) COMMENT 'users, transactions, blocks, etc',
    target_id VARCHAR(255),
    target_name VARCHAR(255),
    old_values JSON COMMENT 'Previous values for auditing',
    new_values JSON COMMENT 'Updated values for auditing',
    changes_summary VARCHAR(500),
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success' COMMENT 'success, failed, pending',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE,
    INDEX idx_admin_id (admin_id),
    INDEX idx_action (action),
    INDEX idx_target_entity (target_entity),
    INDEX idx_created_at (created_at),
    INDEX idx_admin_created (admin_id, created_at)
);

-- Default permissions for each role
-- admin: all permissions
-- moderator: read, moderate_users, moderate_transactions
-- support: read, user_support

-- Insert system admin account (optional)
INSERT INTO admins (user_id, role, permissions, status)
SELECT id, 'admin', JSON_ARRAY('*'), 'active'
FROM users
WHERE address = 'ADMIN_ACCOUNT' AND NOT EXISTS (
    SELECT 1 FROM admins WHERE role = 'admin' LIMIT 1
)
LIMIT 1;
