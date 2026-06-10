package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
)

type NotificationRepository interface {
	Save(ctx context.Context, event dto.NotificationEvent) error
	GetByRecipient(ctx context.Context, address string, limit, offset int) ([]dto.NotificationEvent, error)
	GetUnreadCount(ctx context.Context, address string) (int, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, address string) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type notificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Save(ctx context.Context, event dto.NotificationEvent) error {
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		dataJSON = []byte("{}")
	}

	channelsJSON, err := json.Marshal(event.Channels)
	if err != nil {
		channelsJSON = []byte("[]")
	}

	query := `INSERT INTO notification_events
		(id, type, priority, recipient_address, title, message, data,
		 related_tx_id, related_block_id, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, FALSE, FROM_UNIXTIME(?))`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		string(event.Type),
		string(event.Priority),
		event.RecipientAddress,
		event.Title,
		event.Message,
		dataJSON,
		event.RelatedTxID,
		event.RelatedBlockID,
		event.CreatedAt,
	)

	// Simpan channels sebagai JSON di kolom data jika perlu referensi.
	// Untuk saat ini channels tidak disimpan kolom terpisah karena
	// hanya ws yang dipakai. Bisa ditambah kolom channels jika butuh.
	_ = channelsJSON

	return err
}

func (r *notificationRepository) GetByRecipient(ctx context.Context, address string, limit, offset int) ([]dto.NotificationEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT id, type, priority, recipient_address, title, message,
		data, related_tx_id, related_block_id, is_read, created_at
		FROM notification_events
		WHERE recipient_address = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, address, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []dto.NotificationEvent
	for rows.Next() {
		var (
			id, typ, priority, recipient, title string
			msg                                 sql.NullString
			dataJSON                            []byte
			relatedTxID, relatedBlockID         sql.NullInt64
			isRead                              bool
			createdAt                           time.Time
		)

		if err := rows.Scan(&id, &typ, &priority, &recipient, &title, &msg,
			&dataJSON, &relatedTxID, &relatedBlockID, &isRead, &createdAt); err != nil {
			continue
		}

		event := dto.NotificationEvent{
			ID:               id,
			Type:             dto.NotificationType(typ),
			Priority:         dto.NotificationPriority(priority),
			RecipientAddress: recipient,
			Title:            title,
			Message:          msg.String,
			IsRead:           isRead,
			CreatedAt:        createdAt.Unix(),
		}

		if relatedTxID.Valid {
			v := relatedTxID.Int64
			event.RelatedTxID = &v
		}
		if relatedBlockID.Valid {
			v := relatedBlockID.Int64
			event.RelatedBlockID = &v
		}

		if len(dataJSON) > 0 {
			var data map[string]interface{}
			if json.Unmarshal(dataJSON, &data) == nil {
				event.Data = data
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, address string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM notification_events WHERE recipient_address = ? AND is_read = FALSE`,
		address)
	return count, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notification_events SET is_read = TRUE WHERE id = ?`, id)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, address string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notification_events SET is_read = TRUE WHERE recipient_address = ? AND is_read = FALSE`,
		address)
	return err
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM notification_events WHERE id = ?`, id)
	return err
}

func (r *notificationRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM notification_events WHERE created_at < NOW() - INTERVAL 7 DAY`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
