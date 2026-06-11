package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

type NotificationHandler struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationHandler(notificationRepo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{notificationRepo: notificationRepo}
}

// GetNotifications mengembalikan notifikasi untuk user tertentu.
// Endpoint: GET /notifications?limit=20&offset=0
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("address is required"))
		return
	}

	limit := 20
	offset := 0

	if l, err := parseIntParam(c, "limit"); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := parseIntParam(c, "offset"); err == nil && o >= 0 {
		offset = o
	}

	events, err := h.notificationRepo.GetByRecipient(c.Request.Context(), address, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to get notifications"))
		return
	}

	unreadCount, _ := h.notificationRepo.GetUnreadCount(c.Request.Context(), address)

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{
		"notifications": events,
		"unread_count":  unreadCount,
		"limit":         limit,
		"offset":        offset,
	}))
}

// MarkAsRead menandai satu notifikasi sebagai sudah dibaca.
// Endpoint: PUT /notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("notification id is required"))
		return
	}

	if err := h.notificationRepo.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to mark as read"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{"id": id, "is_read": true}))
}

// MarkAllAsRead menandai semua notifikasi sebagai sudah dibaca.
// Endpoint: PUT /notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("address is required"))
		return
	}

	if err := h.notificationRepo.MarkAllAsRead(c.Request.Context(), address); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to mark all as read"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{"marked_all_read": true}))
}

// DeleteNotification menghapus satu notifikasi.
// Endpoint: DELETE /notifications/:id
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("notification id is required"))
		return
	}

	if err := h.notificationRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to delete notification"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{"deleted": true}))
}

func parseIntParam(c *gin.Context, key string) (int, error) {
	v := c.Query(key)
	if v == "" {
		return 0, nil
	}
	var result int
	_, err := fmt.Sscan(v, &result)
	return result, err
}
