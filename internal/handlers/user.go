package handlers

import (
	"net/http"
	"sjek/internal/database"
	"sjek/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserResponse struct {
	ID            uuid.UUID         `json:"id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	Type          models.UserType   `json:"type"`
	Status        models.UserStatus `json:"status"`
	ActivatedDate *time.Time        `json:"activated_date,omitempty"`
	InactiveDate  *time.Time        `json:"inactive_date,omitempty"`
	Roles         []string          `json:"roles"`
}

// Struct untuk update user
type UpdateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password,omitempty"` // Optional untuk update
}

// Update GetUsers function untuk include field baru
// @Summary      Get all users with pagination and filters
// @Description  Get all users with pagination, filter by type, email contains, username contains
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page     query     int     false  "Page number (default: 1)"
// @Param        limit    query     int     false  "Items per page (default: 10, max: 100)"
// @Param        type     query     string  false  "Filter by user type (ADMIN/DRIVER)"
// @Param        email    query     string  false  "Filter by email contains"
// @Param        username query     string  false  "Filter by username contains"
// @Success      200  {object}  models.PaginatedResponse{data=[]UserResponse}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users [get]
func GetUsers(c *gin.Context) {
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
	typeFilter := c.Query("type")
	emailFilter := c.Query("email")
	usernameFilter := c.Query("username")

	// Validasi input
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 10
	}
	if pagination.Limit > 100 {
		pagination.Limit = 100 // Batasi maksimal 100 item per page
	}

	// Hitung offset
	pagination.Offset = (pagination.Page - 1) * pagination.Limit

	// Build query dengan filter
	query := database.DB.Model(&models.User{})

	// Apply filters
	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	if emailFilter != "" {
		query = query.Where("email ILIKE ?", "%"+emailFilter+"%")
	}
	if usernameFilter != "" {
		query = query.Where("username ILIKE ?", "%"+usernameFilter+"%")
	}

	// Hitung total records dengan filter
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}
	pagination.Total = total

	// Ambil data dengan pagination dan filter
	var users []models.User
	result := query.Preload("Roles").Offset(pagination.Offset).Limit(pagination.Limit).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Transform ke response format
	var response []UserResponse
	for _, user := range users {
		var roles []string
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
		response = append(response, UserResponse{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			Type:          user.Type,
			Status:        user.Status,
			ActivatedDate: user.ActivatedDate,
			InactiveDate:  user.InactiveDate,
			Roles:         roles,
		})
	}

	// Return response
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       response,
		Pagination: pagination,
	})
}

// @Summary      Get user by ID
// @Description  Get user details by user ID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{id} [get]
func GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	result := database.DB.Preload("Roles").First(&user, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Roles:    roles,
	})
}

// @Summary      Delete user
// @Description  Delete user by ID
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result := database.DB.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// @Summary      Update user
// @Description  Update user details by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    string            true  "User ID"
// @Param        request body    UpdateUserRequest true  "User details"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id} [put]
func UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	result := database.DB.First(&user, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update username dan email
	user.Username = req.Username
	user.Email = req.Email

	// Update password hanya jika diberikan
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	result = database.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}
