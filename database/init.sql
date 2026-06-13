-- ============================================================================
-- Initial database schema untuk MySQL 8.0
-- File ini otomatis dijalankan oleh MySQL Docker image saat pertama kali
-- container dibuat (via /docker-entrypoint-initdb.d/).
--
-- Digenerate dari file-file di database/migrations/ dengan semua ALTER TABLE
-- sudah di-resolve menjadi final schema.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. users (harus dibuat pertama karena table lain reference)
-- ----------------------------------------------------------------------------
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(20) DEFAULT 'user' NOT NULL,
    public_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 2. transactions (di-reference oleh ledger, block_transactions)
-- ----------------------------------------------------------------------------
CREATE TABLE transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    fee DECIMAL(20, 8) NOT NULL DEFAULT 0.00100000,
    type ENUM('TRANSFER', 'BUY', 'SELL') NOT NULL DEFAULT 'TRANSFER',
    signature TEXT NOT NULL,
    status ENUM('PENDING', 'SUCCESS', 'FAILED', 'CONFIRMED') DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_from_address (from_address),
    INDEX idx_to_address (to_address),
    INDEX idx_transactions_fee (fee DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 3. blocks
-- ----------------------------------------------------------------------------
CREATE TABLE blocks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    block_number INT NOT NULL,
    previous_hash VARCHAR(255),
    current_hash VARCHAR(255) NOT NULL,
    nonce BIGINT DEFAULT 0,
    difficulty INT DEFAULT 4,
    timestamp BIGINT,
    merkle_root VARCHAR(64),
    miner_address VARCHAR(255),
    block_reward DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    total_fees DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    created_at DATETIME DEFAULT NOW(),

    INDEX idx_blocks_timestamp (timestamp),
    INDEX idx_blocks_difficulty (difficulty),
    INDEX idx_miner_address (miner_address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 4. ledger (FK ke transactions + blocks)
-- ----------------------------------------------------------------------------
CREATE TABLE ledger (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    block_id BIGINT NOT NULL,
    tx_id BIGINT NULL,
    address VARCHAR(255) NOT NULL,
    change_amount DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_block_id (block_id),
    FOREIGN KEY (tx_id) REFERENCES transactions(id),
    FOREIGN KEY (block_id) REFERENCES blocks(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 5. block_transactions (FK ke blocks + transactions)
-- ----------------------------------------------------------------------------
CREATE TABLE block_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    block_id BIGINT NOT NULL,
    transaction_id BIGINT NOT NULL,

    FOREIGN KEY (block_id) REFERENCES blocks(id),
    FOREIGN KEY (transaction_id) REFERENCES transactions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 6. user_wallets (FK ke users.address)
-- ----------------------------------------------------------------------------
CREATE TABLE user_wallets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) UNIQUE NOT NULL,
    yte_balance DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    locked_balance DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    available_balance DECIMAL(20, 8) GENERATED ALWAYS AS (yte_balance - locked_balance) STORED,
    total_received DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    total_sent DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    last_transaction_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_last_transaction (last_transaction_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE,

    CHECK (locked_balance <= yte_balance),
    CHECK (yte_balance >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 7. user_balances (FK ke users.address)
-- ----------------------------------------------------------------------------
CREATE TABLE user_balances (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) UNIQUE NOT NULL,
    usd_balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    locked_balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    available_balance DECIMAL(20, 2) GENERATED ALWAYS AS (usd_balance - locked_balance) STORED,
    total_deposited DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    total_withdrawn DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    total_traded DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    last_transaction_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_last_transaction (last_transaction_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE,

    CHECK (locked_balance <= usd_balance),
    CHECK (usd_balance >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 8. user_balance_history (FK ke users.id)
-- ----------------------------------------------------------------------------
CREATE TABLE user_balance_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    old_balance DECIMAL(18,2),
    new_balance DECIMAL(18,2),
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 9. balance_history (FK ke users.address)
-- ----------------------------------------------------------------------------
CREATE TABLE balance_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) NOT NULL,
    order_id BIGINT NULL,
    change_type ENUM(
        'DEPOSIT', 'WITHDRAWAL', 'BUY_ORDER', 'SELL_ORDER',
        'CANCEL_ORDER', 'FEE', 'LOCK', 'UNLOCK'
    ) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    balance_before DECIMAL(20, 2) NOT NULL,
    balance_after DECIMAL(20, 2) NOT NULL,
    locked_before DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    locked_after DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    reference_id VARCHAR(255) NULL,
    description VARCHAR(512) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_order_id (order_id),
    INDEX idx_change_type (change_type),
    INDEX idx_reference_id (reference_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 10. balance_locks (FK ke users.address)
-- ----------------------------------------------------------------------------
CREATE TABLE balance_locks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    lock_type ENUM('BUY_ORDER', 'SELL_ORDER', 'OTHER') NOT NULL,
    reference_id VARCHAR(255) NOT NULL,
    status ENUM('ACTIVE', 'RELEASED', 'EXECUTED') NOT NULL DEFAULT 'ACTIVE',
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP NULL,

    INDEX idx_user_address (user_address),
    INDEX idx_reference_id (reference_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE,

    UNIQUE KEY unique_lock (reference_id, lock_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 11. wallet_history (FK ke users.address)
-- ----------------------------------------------------------------------------
CREATE TABLE wallet_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) NOT NULL,
    tx_id BIGINT NULL,
    order_id BIGINT NULL,
    change_type ENUM(
        'RECEIVE', 'SEND', 'BUY_ORDER', 'SELL_ORDER',
        'MINING', 'FEE_PAID', 'LOCK', 'UNLOCK'
    ) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    balance_before DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    locked_before DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    locked_after DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    reference_id VARCHAR(255) NULL,
    description VARCHAR(512) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_tx_id (tx_id),
    INDEX idx_order_id (order_id),
    INDEX idx_change_type (change_type),
    INDEX idx_reference_id (reference_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 12. market_engine
-- ----------------------------------------------------------------------------
CREATE TABLE market_engine (
    id INT PRIMARY KEY CHECK (id = 1),
    price DECIMAL(20, 8) NOT NULL,
    liquidity DECIMAL(20, 8) NOT NULL,
    last_block BIGINT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 13. market_ticks (FK ke blocks)
-- ----------------------------------------------------------------------------
CREATE TABLE market_ticks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    block_id BIGINT NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    buy_volume DECIMAL(20, 8) NOT NULL,
    sell_volume DECIMAL(20, 8) NOT NULL,
    tx_count INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_tick_block (block_id),
    FOREIGN KEY (block_id) REFERENCES blocks(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 14. candles
-- ----------------------------------------------------------------------------
CREATE TABLE candles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    interval_type ENUM('1m', '5m', '15m', '30m', '1h', '4h', '1d') NOT NULL,
    start_time BIGINT NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    high_price DECIMAL(20, 8) NOT NULL,
    low_price DECIMAL(20, 8) NOT NULL,
    close_price DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,

    UNIQUE KEY uniq_candle (interval_type, start_time),
    INDEX idx_candle_interval (interval_type),
    INDEX idx_candle_start_time (start_time),
    INDEX idx_candle_interval_time (interval_type, start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 15. staking_records
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS staking_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(42) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    lock_until BIGINT NOT NULL,
    reward_earned DECIMAL(20, 8) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_status (user_address, status),
    INDEX idx_lock_until (lock_until),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 16. staking_rewards
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS staking_rewards (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    stake_id BIGINT NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    reward_amount DECIMAL(20, 8) NOT NULL,
    block_number BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_stake_id (stake_id),
    INDEX idx_user_address (user_address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 17. admins (FK ke users.id)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admins (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    password_hash VARCHAR(255) NOT NULL DEFAULT '',
    permissions JSON,
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    last_login_at TIMESTAMP NULL,
    last_password_changed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_role (role),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 18. admin_activity_logs (FK ke admins.id)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admin_activity_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    admin_id INT NOT NULL,
    action VARCHAR(100) NOT NULL,
    target_entity VARCHAR(50),
    target_id VARCHAR(255),
    target_name VARCHAR(255),
    old_values JSON,
    new_values JSON,
    changes_summary VARCHAR(500),
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE,
    INDEX idx_admin_id (admin_id),
    INDEX idx_action (action),
    INDEX idx_target_entity (target_entity),
    INDEX idx_created_at (created_at),
    INDEX idx_admin_created (admin_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 19. notification_events
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notification_events (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'low',
    recipient_address VARCHAR(42) NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT,
    data JSON,
    related_tx_id BIGINT,
    related_block_id BIGINT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_recipient_read (recipient_address, is_read, created_at DESC),
    INDEX idx_type (type),
    INDEX idx_created (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 20. balance_discrepancy
-- ----------------------------------------------------------------------------
CREATE TABLE balance_discrepancy (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    block_number INT NOT NULL,
    expected_balance DECIMAL(20, 8) NOT NULL,
    actual_balance DECIMAL(20, 8) NOT NULL,
    difference DECIMAL(20, 8) NOT NULL,
    resolved BOOLEAN DEFAULT FALSE,
    resolution_note VARCHAR(512),
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_address (address),
    INDEX idx_block_number (block_number),
    INDEX idx_resolved (resolved),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- SEED DATA: System accounts
-- ============================================================================

-- System account: FEE_POOL
INSERT INTO users(name, address, public_key, role)
VALUES ('FEE_POOL', 'FEE_POOL', 'SYSTEM_ACCOUNT', 'system');

-- System account: MINER_ACCOUNT
INSERT INTO users (name, address, public_key, role)
VALUES ('MINER_ACCOUNT', 'MINER_ACCOUNT', 'SYSTEM_MINER', 'system');

-- Admin account
INSERT INTO users (name, address, public_key, role)
VALUES ('admin', 'ADMIN_ACCOUNT', 'ADMIN_KEY', 'admin');

INSERT INTO admins (user_id, role, permissions, status, password_hash, created_at)
SELECT id, 'admin', JSON_ARRAY('*'), 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeekeJQbIjk', NOW()
FROM users
WHERE name = 'admin' AND role = 'admin'
LIMIT 1;

-- Wallet entries untuk system accounts
INSERT INTO user_wallets (user_address, yte_balance)
VALUES ('FEE_POOL', 0.00000000),
       ('MINER_ACCOUNT', 0.00000000),
       ('ADMIN_ACCOUNT', 0.00000000);

-- Balance entries untuk system accounts
INSERT INTO user_balances (user_address, usd_balance)
VALUES ('FEE_POOL', 0.00),
       ('MINER_ACCOUNT', 0.00),
       ('ADMIN_ACCOUNT', 0.00);

-- Genesis block
INSERT INTO blocks (block_number, previous_hash, current_hash, nonce, difficulty, timestamp, merkle_root, miner_address, block_reward, total_fees)
VALUES (1, '0', '5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9', 0, 4, UNIX_TIMESTAMP(), '', 'GENESIS', 0, 0);

-- Market engine initial state
INSERT INTO market_engine (id, price, liquidity, last_block)
VALUES (1, 1.00000000, 1000000.00000000, 1);
