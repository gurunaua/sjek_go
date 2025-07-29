package database

import (
	"log"
	"sjek/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDefaultMenus creates default menu structure
func SeedDefaultMenus(db *gorm.DB) error {
	// Check if menus already exist
	var count int64
	db.Model(&models.Menu{}).Count(&count)
	if count > 0 {
		log.Println("Menus already exist, skipping seeding")
		return nil
	}

	// Create default menus
	menus := []models.Menu{
		{
			ID:          uuid.New(),
			Name:        "Dashboard",
			Path:        "/dashboard",
			Icon:        "dashboard",
			Sequence:    1,
			IsActive:    true,
			Description: "Main dashboard",
		},
		{
			ID:          uuid.New(),
			Name:        "User Management",
			Path:        "/users",
			Icon:        "users",
			Sequence:    2,
			IsActive:    true,
			Description: "Manage users",
		},
		{
			ID:          uuid.New(),
			Name:        "Role Management",
			Path:        "/roles",
			Icon:        "shield",
			Sequence:    3,
			IsActive:    true,
			Description: "Manage roles",
		},
		{
			ID:          uuid.New(),
			Name:        "API Management",
			Path:        "/apis",
			Icon:        "code",
			Sequence:    4,
			IsActive:    true,
			Description: "Manage APIs",
		},
		{
			ID:          uuid.New(),
			Name:        "Menu Management",
			Path:        "/menus",
			Icon:        "menu",
			Sequence:    5,
			IsActive:    true,
			Description: "Manage menus",
		},
		{
			ID:          uuid.New(),
			Name:        "Login Logs",
			Path:        "/login-logs",
			Icon:        "history",
			Sequence:    6,
			IsActive:    true,
			Description: "View login logs",
		},
		{
			ID:          uuid.New(),
			Name:        "Token Management",
			Path:        "/tokens",
			Icon:        "key",
			Sequence:    7,
			IsActive:    true,
			Description: "Manage tokens",
		},
	}

	// Create submenus for User Management
	userManagementID := menus[1].ID
	userSubmenus := []models.Menu{
		{
			ID:          uuid.New(),
			Name:        "View Users",
			Path:        "/users/list",
			Icon:        "list",
			ParentID:    &userManagementID,
			Sequence:    1,
			IsActive:    true,
			Description: "View all users",
		},
		{
			ID:          uuid.New(),
			Name:        "Add User",
			Path:        "/users/add",
			Icon:        "plus",
			ParentID:    &userManagementID,
			Sequence:    2,
			IsActive:    true,
			Description: "Add new user",
		},
	}

	// Insert main menus
	for _, menu := range menus {
		if err := db.Create(&menu).Error; err != nil {
			return err
		}
	}

	// Insert submenus
	for _, submenu := range userSubmenus {
		if err := db.Create(&submenu).Error; err != nil {
			return err
		}
	}

	// Assign all menus to super_admin role
	var superAdminRole models.Role
	if err := db.Where("name = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		log.Printf("Warning: super_admin role not found, skipping menu assignment")
		return nil
	}

	// Get all created menus
	var allMenus []models.Menu
	db.Find(&allMenus)

	// Assign all menus to super_admin
	for _, menu := range allMenus {
		if err := db.Model(&menu).Association("Roles").Append(&superAdminRole); err != nil {
			log.Printf("Warning: failed to assign menu %s to super_admin: %v", menu.Name, err)
		}
	}

	log.Println("Default menus seeded successfully")
	return nil
}