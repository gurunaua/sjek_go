package routes

import (
	"fmt"
	"log"
	"sjek/internal/handlers"
	"sjek/internal/middleware"
	"sjek/internal/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	setupPublicRoutes(router)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.Use(middleware.APIAccessMiddleware(db))

	setupUserRoutes(protected)
	setupRoleRoutes(protected)
	setupRoleUserMappingRoutes(protected)
	setupAPIRoutes(protected, db)
	setupAPIRoleMappingRoutes(protected, db)
	setupLoginLogRoutes(protected)
	setupTokenRoutes(protected)
	setupMenuRoutes(protected) // Tambahkan ini

	// Insert API endpoints ke database
	if err := insertAPIEndpoints(db, router); err != nil {
		log.Printf("Warning: gagal insert API endpoints: %v", err)
	}

	return router
}

// setupPublicRoutes configures public routes that don't require authentication
func setupPublicRoutes(router *gin.Engine) {
	router.POST("/register/admin", handlers.RegisterAdmin)
	router.POST("/register/driver", handlers.RegisterDriver)
	router.POST("/login", handlers.Login)
}

// setupUserRoutes configures user management routes
func setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("/", handlers.GetUsers)
		users.GET("/:id", handlers.GetUser)
		users.PUT("/:id", handlers.UpdateUser)
		users.DELETE("/:id", handlers.DeleteUser)
	}
}

// setupRoleRoutes configures role management routes
func setupRoleRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	{
		roles.POST("/", handlers.CreateRole)
		roles.GET("/", handlers.GetRoles)
		roles.GET("/:id", handlers.GetRole)
		roles.PUT("/:id", handlers.UpdateRole)
		roles.DELETE("/:id", handlers.DeleteRole)
	}
}

// setupRoleUserMappingRoutes configures role-user mapping routes
func setupRoleUserMappingRoutes(rg *gin.RouterGroup) {
	mappings := rg.Group("/role-assignments")
	{
		mappings.POST("/roles/:role_id/users/:user_id", handlers.AssignRoleToUser)
		mappings.DELETE("/roles/:role_id/users/:user_id", handlers.RemoveRoleFromUser)
	}
}

// setupAPIRoutes configures API management routes
func setupAPIRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	apis := rg.Group("/apis")
	{
		apis.POST("/", handlers.CreateAPI(db))
		apis.GET("/", handlers.GetAPIs(db))
		apis.GET("/:id", handlers.GetAPI(db))
		apis.PUT("/:id", handlers.UpdateAPI(db))
		apis.DELETE("/:id", handlers.DeleteAPI(db))
	}
}

// setupAPIRoleMappingRoutes configures API-role mapping routes
func setupAPIRoleMappingRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	mappings := rg.Group("/api-assignments")
	{
		mappings.POST("/apis/:api_id/roles/:role_id", handlers.AssignAPIToRole(db))
		mappings.DELETE("/apis/:api_id/roles/:role_id", handlers.RemoveAPIFromRole(db))
	}
}

// insertAPIEndpoints inserts all registered routes into the database
func insertAPIEndpoints(db *gorm.DB, router *gin.Engine) error {
	// Mendapatkan semua routes yang terdaftar
	for _, routeInfo := range router.Routes() {
		// Membuat API entry baru
		api := models.API{
			Path:        routeInfo.Path,
			Method:      routeInfo.Method,
			Description: fmt.Sprintf("%s %s endpoint", routeInfo.Method, routeInfo.Path),
		}

		// Cek apakah API sudah ada
		var existingAPI models.API
		result := db.Where("path = ? AND method = ?", api.Path, api.Method).First(&existingAPI)
		if result.Error == gorm.ErrRecordNotFound {
			// Insert API baru jika belum ada
			if err := db.Create(&api).Error; err != nil {
				return fmt.Errorf("gagal membuat API %s %s: %v", api.Method, api.Path, err)
			}
			log.Printf("API berhasil dibuat: %s %s", api.Method, api.Path)
		}
	}

	return nil
}

// setupLoginLogRoutes configures login log routes
func setupLoginLogRoutes(rg *gin.RouterGroup) {
	loginLogs := rg.Group("/login-logs")
	{
		loginLogs.GET("/", handlers.GetLoginLogs)
		loginLogs.GET("/:id", handlers.GetLoginLog)
	}
}

// setupTokenRoutes configures token management routes
func setupTokenRoutes(rg *gin.RouterGroup) {
	// Logout route
	rg.POST("/logout", handlers.Logout)
	
	// Token management routes
	tokens := rg.Group("/tokens")
	{
		tokens.GET("/", handlers.GetUserTokens)
		tokens.DELETE("/:id", handlers.RevokeToken)
		tokens.POST("/revoke-all", handlers.RevokeAllTokens)
	}
}

// setupMenuRoutes configures menu management routes
func setupMenuRoutes(rg *gin.RouterGroup) {
	// Menu CRUD routes
	menus := rg.Group("/menus")
	{
		menus.POST("/", handlers.CreateMenu)
		menus.GET("/", handlers.GetMenus)
		menus.GET("/user", handlers.GetUserMenus) // Get menus for current user
		menus.GET("/:id", handlers.GetMenu)
		menus.PUT("/:id", handlers.UpdateMenu)
		menus.DELETE("/:id", handlers.DeleteMenu)
	}

	// Menu-role assignment routes
	menuAssignments := rg.Group("/menu-assignments")
	{
		menuAssignments.POST("/menus/:menu_id/roles/:role_id", handlers.AssignRoleToMenu)
		menuAssignments.DELETE("/menus/:menu_id/roles/:role_id", handlers.RemoveRoleFromMenu)
	}
}
