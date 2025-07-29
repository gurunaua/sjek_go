package handlers

import (
	"net/http"
	"sjek/internal/database"
	"sjek/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoginLogResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	LoginTime time.Time `json:"login_time"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
}

// @Summary      Get login logs with pagination and filters
// @Description  Get all login logs with pagination, filter by status, username, date range
// @Tags         login-logs
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        limit     query     int     false  "Items per page (default: 10, max: 100)"
// @Param        status    query     string  false  "Filter by status (SUCCESS/FAILED)"
// @Param        username  query     string  false  "Filter by username contains"
// @Param        from_date query     string  false  "Filter from date (YYYY-MM-DD)"
// @Param        to_date   query     string  false  "Filter to date (YYYY-MM-DD)"
// @Success      200  {object}  models.PaginatedResponse{data=[]LoginLogResponse}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /login-logs [get]
func GetLoginLogs(c *gin.Context) {
	var pagination models.Pagination
	// Set default values
	pagination.Page = 1
	pagination.Limit = 10

	// Bind query parameters
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get filter parameters
	statusFilter := c.Query("status")
	usernameFilter := c.Query("username")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	// Validasi input
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 10
	}
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}

	// Hitung offset
	pagination.Offset = (pagination.Page - 1) * pagination.Limit

	// Build query dengan filter
	query := database.DB.Model(&models.LoginLog{})

	// Apply filters
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if usernameFilter != "" {
		query = query.Where("username ILIKE ?", "%"+usernameFilter+"%")
	}
	if fromDate != "" {
		query = query.Where("login_time >= ?", fromDate+" 00:00:00")
	}
	if toDate != "" {
		query = query.Where("login_time <= ?", toDate+" 23:59:59")
	}

	// Hitung total records dengan filter
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count login logs"})
		return
	}
	pagination.Total = total

	// Ambil data dengan pagination dan filter, order by login_time desc
	var loginLogs []models.LoginLog
	result := query.Order("login_time DESC").Offset(pagination.Offset).Limit(pagination.Limit).Find(&loginLogs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch login logs"})
		return
	}

	// Transform ke response format
	var response []LoginLogResponse
	for _, log := range loginLogs {
		response = append(response, LoginLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Username:  log.Username,
			Email:     log.Email,
			IPAddress: log.IPAddress,
			UserAgent: log.UserAgent,
			LoginTime: log.LoginTime,
			Status:    log.Status,
			Message:   log.Message,
		})
	}

	// Return response
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       response,
		Pagination: pagination,
	})
}

// @Summary      Get login log by ID
// @Description  Get login log details by ID
// @Tags         login-logs
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Login Log ID"
// @Success      200  {object}  LoginLogResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /login-logs/{id} [get]
func GetLoginLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login log ID"})
		return
	}

	var loginLog models.LoginLog
	result := database.DB.First(&loginLog, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Login log not found"})
		return
	}

	c.JSON(http.StatusOK, LoginLogResponse{
		ID:        loginLog.ID,
		UserID:    loginLog.UserID,
		Username:  loginLog.Username,
		Email:     loginLog.Email,
		IPAddress: loginLog.IPAddress,
		UserAgent: loginLog.UserAgent,
		LoginTime: loginLog.LoginTime,
		Status:    loginLog.Status,
		Message:   loginLog.Message,
	})
}