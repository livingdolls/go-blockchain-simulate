-- Staking records table untuk Proof-of-Stake simulation.
-- User bisa stake YTE untuk earn rewards.
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
);

-- Staking rewards log untuk audit trail.
CREATE TABLE IF NOT EXISTS staking_rewards (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    stake_id BIGINT NOT NULL,
    user_address VARCHAR(42) NOT NULL,
    reward_amount DECIMAL(20, 8) NOT NULL,
    block_number BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_stake_id (stake_id),
    INDEX idx_user_address (user_address)
);
