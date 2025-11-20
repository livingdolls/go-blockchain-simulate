CREATE TABLE ledger (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    tx_id BIGINT NOT NULL,
    address VARCHAR(255) NOT NULL,
    change_amount DECIMAL(18,2) NOT NULL,
    balance_after DECIMAL(18,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (tx_id) REFERENCES transactions(id)
);
