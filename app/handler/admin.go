package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

// parseAdminID mem-parse ID dari URL path parameter dengan validasi.
// Return 0 dan false jika ID invalid (bukan angka atau ≤ 0).
// Sebelumnya: strconv.Atoi(c.Param("id")) error discards → ID=0 silently.
func parseAdminID(c *gin.Context) (int, bool) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid admin ID"))
		return 0, false
	}
	return id, true
}

// parseLimitOffset mem-parse limit dan offset dari query parameters
// dengan validasi range. Return defaults jika tidak di-set.
// Sebelumnya: strconv.Atoi error discards → limit/offset negatif lolos.
func parseLimitOffset(c *gin.Context, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	offset = 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= maxLimit {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	return limit, offset
}

type AdminHandler struct {
	service services.AdminService
}

func NewAdminHandler(service services.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	admin, err := GetAdminFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	ctx := c.Request.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stats, err := h.service.GetDashboardStats(ctx, admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(stats))
}

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	admin, err := GetAdminFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	limit, offset := parseLimitOffset(c, 10, 100)

	admins, err := h.service.GetAllAdmins(ctx, admin, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(admins))
}

func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	admin, err := GetAdminFromContext(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	var req struct {
		UserID      int      `json:"user_id" binding:"required,gt=0"`
		Role        string   `json:"role" binding:"required,oneof=admin moderator support"`
		Permissions []string `json:"permissions"`
	}

	if !dto.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = h.service.CreateAdmin(ctx, admin, req.UserID, req.Role, req.Permissions)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Admin created successfully"))
}

func (h *AdminHandler) UpdateAdminRole(c *gin.Context) {
	admin, err := GetAdminFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	targetAdminID, ok := parseAdminID(c)
	if !ok {
		return
	}

	var req struct {
		Role        string   `json:"role" binding:"required,oneof=admin moderator support"`
		Permissions []string `json:"permissions"`
	}

	if !dto.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = h.service.UpdateAdminRole(ctx, admin, targetAdminID, req.Role, req.Permissions)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Admin role updated successfully"))
}

func (h *AdminHandler) UpdateAdminStatus(c *gin.Context) {
	admin, err := GetAdminFromContext(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	targetAdminID, ok := parseAdminID(c)
	if !ok {
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if !dto.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = h.service.UpdateAdminStatus(ctx, admin, targetAdminID, req.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Admin status updated successfully"))
}

func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	admin, err := GetAdminFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	targetAdminID, ok := parseAdminID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = h.service.DeleteAdmin(ctx, admin, targetAdminID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Admin deleted successfully"))
}

func (h *AdminHandler) GetActivityLogs(c *gin.Context) {
	admin, err := GetAdminFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	targetAdminID := 0
	if aid := c.Query("admin_id"); aid != "" {
		if v, err := strconv.Atoi(aid); err == nil && v > 0 {
			targetAdminID = v
		}
	}

	action := c.Query("action")
	limit, offset := parseLimitOffset(c, 50, 200)

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logs, err := h.service.GetActivityLogs(ctx, admin, targetAdminID, action, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(logs))
}

func (h *AdminHandler) RecentActivityLogs(c *gin.Context) {
	admin, err := GetAdminFromContext(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: admin not found"))
		return
	}

	days := 7
	limit := 100

	if d := c.Query("days"); d != "" {
		days, _ = strconv.Atoi(d)
	}

	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logs, err := h.service.GetRecentActivityLogs(ctx, admin, days, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(logs))
}
