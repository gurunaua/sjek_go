package handlers

import (
	"net/http"
	"sjek/internal/database"
	"sjek/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TokenResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// @Summary      Logout user
// @Description  Logout user by deactivating current token
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /logout [post]
func Logout(c *gin.Context) {
	tokenID, exists := c.Get("token_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token ID not found"})
		return
	}

	// Parse token ID
	tokenUUID, err := uuid.Parse(tokenID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Deactivate token
	result := database.DB.Model(&models.UserToken{}).Where("id = ?", tokenUUID).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// @Summary      Get user active tokens
// @Description  Get all active tokens for current user
// @Tags         tokens
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   TokenResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tokens [get]
func GetUserTokens(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get active tokens
	var tokens []models.UserToken
	result := database.DB.Where("user_id = ? AND is_active = ? AND expires_at > ?", 
		userUUID, true, time.Now()).Order("created_at DESC").Find(&tokens)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tokens"})
		return
	}

	// Transform to response format
	var response []TokenResponse
	for _, token := range tokens {
		response = append(response, TokenResponse{
			ID:        token.ID,
			UserID:    token.UserID,
			IPAddress: token.IPAddress,
			UserAgent: token.UserAgent,
			ExpiresAt: token.ExpiresAt,
			IsActive:  token.IsActive,
			CreatedAt: token.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Summary      Revoke token
// @Description  Revoke/deactivate a specific token by ID
// @Tags         tokens
// @Security     BearerAuth
// @Param        id   path      string  true  "Token ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tokens/{id} [delete]
func RevokeToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	tokenID := c.Param("id")
	tokenUUID, err := uuid.Parse(tokenID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if token belongs to user
	var token models.UserToken
	result := database.DB.Where("id = ? AND user_id = ?", tokenUUID, userUUID).First(&token)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	// Deactivate token
	result = database.DB.Model(&token).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}

// @Summary      Revoke all tokens
// @Description  Revoke/deactivate all tokens for current user (logout from all devices)
// @Tags         tokens
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tokens/revoke-all [post]
func RevokeAllTokens(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Deactivate all tokens for user
	result := database.DB.Model(&models.UserToken{}).Where("user_id = ?", userUUID).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All tokens revoked successfully"})
}