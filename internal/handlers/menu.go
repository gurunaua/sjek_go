package handlers

import (
	"net/http"
	"sjek/internal/database"
	"sjek/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MenuRequest struct {
	Name        string     `json:"name" binding:"required"`
	Path        string     `json:"path" binding:"required"`
	Icon        string     `json:"icon,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Sequence    int        `json:"sequence"`
	IsActive    bool       `json:"is_active"`
	Description string     `json:"description,omitempty"`
}

type MenuResponse struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Path        string         `json:"path"`
	Icon        string         `json:"icon,omitempty"`
	ParentID    *uuid.UUID     `json:"parent_id,omitempty"`
	Sequence    int            `json:"sequence"`
	IsActive    bool           `json:"is_active"`
	Description string         `json:"description,omitempty"`
	Children    []MenuResponse `json:"children,omitempty"`
	Roles       []string       `json:"roles,omitempty"`
}

// @Summary      Create new menu
// @Description  Create a new menu item
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body MenuRequest true "Menu details"
// @Success      201  {object}  MenuResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menus [post]
func CreateMenu(c *gin.Context) {
	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate parent menu exists if ParentID is provided
	if req.ParentID != nil {
		var parentMenu models.Menu
		if err := database.DB.First(&parentMenu, "id = ?", *req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent menu not found"})
			return
		}
	}

	menu := models.Menu{
		ID:          uuid.New(),
		Name:        req.Name,
		Path:        req.Path,
		Icon:        req.Icon,
		ParentID:    req.ParentID,
		Sequence:    req.Sequence,
		IsActive:    req.IsActive,
		Description: req.Description,
	}

	if err := database.DB.Create(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
		return
	}

	response := MenuResponse{
		ID:          menu.ID,
		Name:        menu.Name,
		Path:        menu.Path,
		Icon:        menu.Icon,
		ParentID:    menu.ParentID,
		Sequence:    menu.Sequence,
		IsActive:    menu.IsActive,
		Description: menu.Description,
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary      Get all menus
// @Description  Get list of all menus with hierarchical structure
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        flat query bool false "Return flat list instead of hierarchical"
// @Success      200  {array}   MenuResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menus [get]
func GetMenus(c *gin.Context) {
	flat := c.Query("flat") == "true"

	var menus []models.Menu
	query := database.DB.Preload("Roles").Order("sequence ASC, name ASC")
	
	if flat {
		// Return all menus in flat structure
		if err := query.Find(&menus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menus"})
			return
		}
	} else {
		// Return only root menus (no parent) with children
		if err := query.Where("parent_id IS NULL").Find(&menus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menus"})
			return
		}
		
		// Load children recursively
		for i := range menus {
			loadMenuChildren(&menus[i])
		}
	}

	var response []MenuResponse
	for _, menu := range menus {
		response = append(response, buildMenuResponse(menu))
	}

	c.JSON(http.StatusOK, response)
}

// @Summary      Get menu by ID
// @Description  Get menu details by menu ID
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {object}  MenuResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /menus/{id} [get]
func GetMenu(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	var menu models.Menu
	if err := database.DB.Preload("Roles").Preload("Children").First(&menu, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menu"})
		}
		return
	}

	response := buildMenuResponse(menu)
	c.JSON(http.StatusOK, response)
}

// @Summary      Update menu
// @Description  Update menu details by ID
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    string      true  "Menu ID"
// @Param        request body    MenuRequest true  "Menu details"
// @Success      200  {object}  MenuResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menus/{id} [put]
func UpdateMenu(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var menu models.Menu
	if err := database.DB.First(&menu, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menu"})
		}
		return
	}

	// Validate parent menu exists if ParentID is provided
	if req.ParentID != nil && *req.ParentID != menu.ID {
		var parentMenu models.Menu
		if err := database.DB.First(&parentMenu, "id = ?", *req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent menu not found"})
			return
		}
		
		// Prevent circular reference
		if isCircularReference(*req.ParentID, menu.ID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Circular reference detected"})
			return
		}
	}

	// Update menu fields
	menu.Name = req.Name
	menu.Path = req.Path
	menu.Icon = req.Icon
	menu.ParentID = req.ParentID
	menu.Sequence = req.Sequence
	menu.IsActive = req.IsActive
	menu.Description = req.Description

	if err := database.DB.Save(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
		return
	}

	response := MenuResponse{
		ID:          menu.ID,
		Name:        menu.Name,
		Path:        menu.Path,
		Icon:        menu.Icon,
		ParentID:    menu.ParentID,
		Sequence:    menu.Sequence,
		IsActive:    menu.IsActive,
		Description: menu.Description,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary      Delete menu
// @Description  Delete menu by ID
// @Tags         menus
// @Security     BearerAuth
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menus/{id} [delete]
func DeleteMenu(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	// Check if menu has children
	var childCount int64
	database.DB.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete menu with children. Delete children first."})
		return
	}

	result := database.DB.Delete(&models.Menu{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu deleted successfully"})
}

// @Summary      Assign role to menu
// @Description  Assign a role to a menu for access control
// @Tags         menu-assignments
// @Security     BearerAuth
// @Param        menu_id path string true "Menu ID"
// @Param        role_id path string true "Role ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menu-assignments/menus/{menu_id}/roles/{role_id} [post]
func AssignRoleToMenu(c *gin.Context) {
	menuID, err := uuid.Parse(c.Param("menu_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// Check if menu exists
	var menu models.Menu
	if err := database.DB.First(&menu, "id = ?", menuID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	// Check if role exists
	var role models.Role
	if err := database.DB.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Assign role to menu
	if err := database.DB.Model(&menu).Association("Roles").Append(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role to menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned to menu successfully"})
}

// @Summary      Remove role from menu
// @Description  Remove a role from a menu
// @Tags         menu-assignments
// @Security     BearerAuth
// @Param        menu_id path string true "Menu ID"
// @Param        role_id path string true "Role ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menu-assignments/menus/{menu_id}/roles/{role_id} [delete]
func RemoveRoleFromMenu(c *gin.Context) {
	menuID, err := uuid.Parse(c.Param("menu_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// Check if menu exists
	var menu models.Menu
	if err := database.DB.First(&menu, "id = ?", menuID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	// Check if role exists
	var role models.Role
	if err := database.DB.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Remove role from menu
	if err := database.DB.Model(&menu).Association("Roles").Delete(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role from menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed from menu successfully"})
}

// @Summary      Get user menus
// @Description  Get menus accessible by current user based on their roles
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   MenuResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /menus/user [get]
func GetUserMenus(c *gin.Context) {
	userRoles, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User roles not found"})
		return
	}

	roles := userRoles.([]string)
	if len(roles) == 0 {
		c.JSON(http.StatusOK, []MenuResponse{})
		return
	}

	// Check if user is super_admin
	isSuperAdmin := false
	for _, role := range roles {
		if role == "super_admin" {
			isSuperAdmin = true
			break
		}
	}

	// Get menus accessible by user roles
	var menus []models.Menu
	query := database.DB.Preload("Roles").
		Joins("JOIN map_role_menu ON menus.id = map_role_menu.menu_id").
		Joins("JOIN roles ON map_role_menu.role_id = roles.id").
		Where("roles.name IN ? AND menus.is_active = ?", roles, true).
		Where("menus.parent_id IS NULL").
		Order("menus.sequence ASC, menus.name ASC").
		Distinct()

	if err := query.Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user menus"})
		return
	}

	// Load children recursively (only accessible ones)
	for i := range menus {
		loadUserMenuChildren(&menus[i], roles)
	}

	var response []MenuResponse
	for _, menu := range menus {
		response = append(response, buildUserMenuResponse(menu, isSuperAdmin))
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions
func loadMenuChildren(menu *models.Menu) {
	database.DB.Preload("Roles").Where("parent_id = ?", menu.ID).Order("sequence ASC, name ASC").Find(&menu.Children)
	for i := range menu.Children {
		loadMenuChildren(&menu.Children[i])
	}
}

func loadUserMenuChildren(menu *models.Menu, userRoles []string) {
	var children []models.Menu
	database.DB.Preload("Roles").
		Joins("JOIN map_role_menu ON menus.id = map_role_menu.menu_id").
		Joins("JOIN roles ON map_role_menu.role_id = roles.id").
		Where("roles.name IN ? AND menus.is_active = ? AND menus.parent_id = ?", userRoles, true, menu.ID).
		Order("menus.sequence ASC, menus.name ASC").
		Distinct().
		Find(&children)

	menu.Children = children
	for i := range menu.Children {
		loadUserMenuChildren(&menu.Children[i], userRoles)
	}
}

func buildMenuResponse(menu models.Menu) MenuResponse {
	var roles []string
	for _, role := range menu.Roles {
		roles = append(roles, role.Name)
	}

	var children []MenuResponse
	for _, child := range menu.Children {
		children = append(children, buildMenuResponse(child))
	}

	return MenuResponse{
		ID:          menu.ID,
		Name:        menu.Name,
		Path:        menu.Path,
		Icon:        menu.Icon,
		ParentID:    menu.ParentID,
		Sequence:    menu.Sequence,
		IsActive:    menu.IsActive,
		Description: menu.Description,
		Children:    children,
		Roles:       roles,
	}
}

// New function for user menu response (conditionally include roles)
func buildUserMenuResponse(menu models.Menu, includeRoles bool) MenuResponse {
	var children []MenuResponse
	for _, child := range menu.Children {
		children = append(children, buildUserMenuResponse(child, includeRoles))
	}

	response := MenuResponse{
		ID:          menu.ID,
		Name:        menu.Name,
		Path:        menu.Path,
		Icon:        menu.Icon,
		ParentID:    menu.ParentID,
		Sequence:    menu.Sequence,
		IsActive:    menu.IsActive,
		Description: menu.Description,
		Children:    children,
	}

	// Only include roles if user is super_admin
	if includeRoles {
		var roles []string
		for _, role := range menu.Roles {
			roles = append(roles, role.Name)
		}
		response.Roles = roles
	}

	return response
}

func isCircularReference(parentID, menuID uuid.UUID) bool {
	var menu models.Menu
	if err := database.DB.First(&menu, "id = ?", parentID).Error; err != nil {
		return false
	}

	if menu.ParentID == nil {
		return false
	}

	if *menu.ParentID == menuID {
		return true
	}

	return isCircularReference(*menu.ParentID, menuID)
}
