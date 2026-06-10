-- Notification events table untuk persist notifikasi user.
-- Dipakai oleh NotificationWSConsumer untuk menyimpan event sebelum
-- push ke WebSocket client. User bisa lihat history setelah refresh.
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
);
