package handlers

import (
	"net/http"
	"sjek/internal/database"
	"sjek/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoleRequest struct {
	Name string `json:"name" binding:"required"`
}

type RoleResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// @Summary      Create new role
// @Description  Create a new role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body RoleRequest true "Role details"
// @Success      201  {object}  RoleResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /roles [post]
func CreateRole(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		ID:   uuid.New(),
		Name: req.Name,
	}

	result := database.DB.Create(&role)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, RoleResponse{ID: role.ID, Name: role.Name})
}

// @Summary      Get all roles
// @Description  Get list of all roles
// @Tags         roles
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   RoleResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /roles [get]
func GetRoles(c *gin.Context) {
	var roles []models.Role
	result := database.DB.Find(&roles)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}

	var response []RoleResponse
	for _, role := range roles {
		response = append(response, RoleResponse{ID: role.ID, Name: role.Name})
	}

	c.JSON(http.StatusOK, response)
}

// @Summary      Get role by ID
// @Description  Get role details by role ID
// @Tags         roles
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  RoleResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /roles/{id} [get]
func GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role models.Role
	result := database.DB.First(&role, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, RoleResponse{ID: role.ID, Name: role.Name})
}

// @Summary      Update role
// @Description  Update role details by ID
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    string      true  "Role ID"
// @Param        request body    RoleRequest true  "Role details"
// @Success      200  {object}  RoleResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /roles/{id} [put]
func UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	result := database.DB.First(&role, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	role.Name = req.Name
	result = database.DB.Save(&role)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, RoleResponse{ID: role.ID, Name: role.Name})
}

// @Summary      Delete role
// @Description  Delete role by ID
// @Tags         roles
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	result := database.DB.Delete(&models.Role{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// @Summary      Assign role to user
// @Description  Assign a role to a user
// @Tags         role-assignments
// @Security     BearerAuth
// @Param        role_id path    string true "Role ID"
// @Param        user_id path    string true "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /role-assignments/roles/{role_id}/users/{user_id} [post]
func AssignRoleToUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var user models.User
	result := database.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var role models.Role
	result = database.DB.First(&role, "id = ?", roleID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	if err := database.DB.Model(&user).Association("Roles").Append(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned to user successfully"})
}

// @Summary      Remove role from user
// @Description  Remove a role from a user
// @Tags         role-assignments
// @Security     BearerAuth
// @Param        role_id path    string true "Role ID"
// @Param        user_id path    string true "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /role-assignments/roles/{role_id}/users/{user_id} [delete]
func RemoveRoleFromUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var user models.User
	result := database.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var role models.Role
	result = database.DB.First(&role, "id = ?", roleID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Perbaikan: Tidak perlu menyimpan hasil Association().Delete() ke dalam result
	if err := database.DB.Model(&user).Association("Roles").Delete(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role from user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed from user successfully"})
}
