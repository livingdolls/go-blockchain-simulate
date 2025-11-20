CREATE TABLE transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    signature TEXT NOT NULL,
    status ENUM('PENDING', 'SUCCESS', 'FAILED', 'CONFIRMED') DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_from_address (from_address),
    INDEX idx_to_address (to_address)
);
