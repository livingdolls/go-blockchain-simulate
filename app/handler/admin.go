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

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	if o := c.Query("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

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
		UserID      int      `json:"user_id" binding:"required"`
		Role        string   `json:"role" binding:"required"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("Invalid request body"))
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

	targetAdminID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Role        string   `json:"role" binding:"required"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("Invalid request body"))
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

	targetAdminID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("Invalid request body"))
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

	targetAdminID, _ := strconv.Atoi(c.Param("id"))

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
		targetAdminID, _ = strconv.Atoi(aid)
	}

	action := c.Query("action")
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	if o := c.Query("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

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
