package routes

import (
	"fmt"
	"log"
	"sjek/internal/handlers"
	"sjek/internal/middleware"
	"sjek/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter configures all the routes for the application
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Public routes
	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// User routes
	userRoutes := protected.Group("/users")
	{
		userRoutes.GET("/", handlers.GetUsers)
		userRoutes.GET("/:id", handlers.GetUser)
		userRoutes.PUT("/:id", handlers.UpdateUser)
		userRoutes.DELETE("/:id", handlers.DeleteUser)
	}

	// Role routes
	roleRoutes := protected.Group("/roles")
	{
		roleRoutes.POST("/", handlers.CreateRole)
		roleRoutes.GET("/", handlers.GetRoles)
		roleRoutes.GET("/:id", handlers.GetRole)
		roleRoutes.PUT("/:id", handlers.UpdateRole)
		roleRoutes.DELETE("/:id", handlers.DeleteRole)
	}

	// Role-User mapping routes
	roleUserRoutes := protected.Group("/role-assignments")
	{
		roleUserRoutes.POST("/roles/:role_id/users/:user_id", handlers.AssignRoleToUser)
		roleUserRoutes.DELETE("/roles/:role_id/users/:user_id", handlers.RemoveRoleFromUser)
	}

	// API routes
	apiRoutes := protected.Group("/apis")
	{
		apiRoutes.POST("/", handlers.CreateAPI(db))
		apiRoutes.GET("/", handlers.GetAPIs(db))
		apiRoutes.GET("/:id", handlers.GetAPI(db))
		apiRoutes.PUT("/:id", handlers.UpdateAPI(db))
		apiRoutes.DELETE("/:id", handlers.DeleteAPI(db))
	}

	// API-Role mapping routes
	apiRoleRoutes := protected.Group("/api-assignments")
	{
		apiRoleRoutes.POST("/apis/:api_id/roles/:role_id", handlers.AssignAPIToRole(db))
		apiRoleRoutes.DELETE("/apis/:api_id/roles/:role_id", handlers.RemoveAPIFromRole(db))
	}

	// Insert API endpoints ke database
	if err := insertAPIEndpoints(db, router); err != nil {
		log.Printf("Warning: gagal insert API endpoints: %v", err)
	}

	return router
}

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
