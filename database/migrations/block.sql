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
