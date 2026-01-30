CREATE TABLE balance_discrepancy (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    block_number  INT NOT NULL,
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
)