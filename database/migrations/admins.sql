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

-- Add password_hash column to admins table for authentication
ALTER TABLE admins 
ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT '' AFTER role,
ADD COLUMN last_password_changed_at TIMESTAMP NULL AFTER last_login_at;

-- Update timestamp column if not exists
ALTER TABLE admins 
MODIFY COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- Create indexes untuk performance
ALTER TABLE admins ADD INDEX idx_user_id (user_id);
ALTER TABLE admins ADD INDEX idx_role (role);
ALTER TABLE admins ADD INDEX idx_status (status);

-- Add constraint untuk ensure password terisi pada admin aktif
-- ALTER TABLE admins ADD CONSTRAINT check_password CHECK (
--   (status = 'active' AND password_hash != '') OR status != 'active'
-- );

-- Pastikan user exist dulu
INSERT INTO users (name, address, public_key, role)
VALUES ('admin', 'ADMIN_ACCOUNT', 'ADMIN_KEY', 'admin')
ON DUPLICATE KEY UPDATE role = 'admin';

-- Baru insert admin (dengan user_id yang exist)
INSERT INTO admins (user_id, role, permissions, status, password_hash, created_at)
SELECT id, 'admin', JSON_ARRAY('*'), 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeekeJQbIjk', NOW()
FROM users 
WHERE name = 'admin' AND role = 'admin' 
LIMIT 1;