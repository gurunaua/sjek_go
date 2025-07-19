package handlers

import (
	"net/http"
	"sjek/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateAPI creates a new API endpoint
func CreateAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var api models.API
		if err := c.ShouldBindJSON(&api); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := db.Create(&api)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusCreated, api)
	}
}

// GetAPIs returns all API endpoints
func GetAPIs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var apis []models.API
		result := db.Preload("Roles").Find(&apis)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, apis)
	}
}

// GetAPI returns a specific API endpoint by ID
func GetAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API ID"})
			return
		}

		var api models.API
		result := db.Preload("Roles").First(&api, "id = ?", id)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		c.JSON(http.StatusOK, api)
	}
}

// UpdateAPI updates an API endpoint
func UpdateAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API ID"})
			return
		}

		var api models.API
		if err := db.First(&api, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		if err := c.ShouldBindJSON(&api); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		api.ID = id // Ensure ID remains unchanged
		if err := db.Save(&api).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, api)
	}
}

// DeleteAPI deletes an API endpoint
func DeleteAPI(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API ID"})
			return
		}

		result := db.Delete(&models.API{}, "id = ?", id)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "API deleted successfully"})
	}
}

// AssignAPIToRole assigns an API endpoint to a role
func AssignAPIToRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiID, err := uuid.Parse(c.Param("api_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API ID"})
			return
		}

		roleID, err := uuid.Parse(c.Param("role_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Role ID"})
			return
		}

		// Check if API exists
		var api models.API
		if err := db.First(&api, "id = ?", apiID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// Check if Role exists
		var role models.Role
		if err := db.First(&role, "id = ?", roleID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}

		// Associate API with Role
		if err := db.Model(&api).Association("Roles").Append(&role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "API assigned to role successfully"})
	}
}

// RemoveAPIFromRole removes an API endpoint from a role
func RemoveAPIFromRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiID, err := uuid.Parse(c.Param("api_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API ID"})
			return
		}

		roleID, err := uuid.Parse(c.Param("role_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Role ID"})
			return
		}

		// Check if API exists
		var api models.API
		if err := db.First(&api, "id = ?", apiID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// Check if Role exists
		var role models.Role
		if err := db.First(&role, "id = ?", roleID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}

		// Remove association between API and Role
		if err := db.Model(&api).Association("Roles").Delete(&role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "API removed from role successfully"})
	}
}
