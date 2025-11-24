CREATE TABLE blocks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    block_number INT NOT NULL,
    previous_hash VARCHAR(255),
    current_hash VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT NOW()
);

CREATE TABLE block_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    block_id BIGINT NOT NULL,
    transaction_id BIGINT NOT NULL,
    FOREIGN KEY (block_id) REFERENCES blocks(id),
    FOREIGN KEY (transaction_id) REFERENCES transactions(id)
);

-- Add PoW fields to block
ALTER TABLE blocks
ADD COLUMN nonce BIGINT DEFAULT 0 AFTER current_hash,
ADD COLUMN difficulty INT DEFAULT 4 AFTER nonce,
ADD COLUMN timestamp BIGINT AFTER difficulty,
ADD COLUMN merkle_root VARCHAR(64) AFTER timestamp;

-- update existing blocks to have timestamp and merkle_root
UPDATE blocks
SET timestamp = UNIX_TIMESTAMP(created_at),
    difficulty = 4,
    nonce = 0
WHERE timestamp IS NULL;

-- Add index
CREATE INDEX idx_blocks_timestamp on blocks (timestamp);
CREATE INDEX idx_blocks_difficulty on blocks (difficulty);

-- add miner_address to blocks
ALTER TABLE blocks
ADD COLUMN miner_address VARCHAR(255) AFTER merkle_root,
ADD COLUMN block_reward DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000 AFTER miner_address,
ADD COLUMN total_fees DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000 AFTER block_reward,
ADD INDEX idx_miner_address (miner_address);