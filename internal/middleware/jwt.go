package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"sjek/internal/models" // perbaiki import path

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, username string, roles []string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := bearerToken[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JWTSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User roles not found"})
			c.Abort()
			return
		}

		roles := userRoles.([]string)
		hasRequiredRole := false
		for _, role := range roles {
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func APIAccessMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dapatkan roles dari context yang sudah diset oleh AuthMiddleware
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User roles not found"})
			c.Abort()
			return
		}
		roles := userRoles.([]string)

		// Dapatkan current path dan method
		path := c.FullPath()
		method := c.Request.Method

		// Cari API di database
		var api models.API
		if err := db.Where("path = ? AND method = ?", path, method).First(&api).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			c.Abort()
			return
		}

		// Cek apakah user memiliki role yang bisa mengakses API ini
		var count int64
		if err := db.Table("map_role_api").Where("api_id = ? AND role_id IN (SELECT id FROM roles WHERE name IN ?)", api.ID, roles).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check API access"})
			c.Abort()
			return
		}

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this API"})
			c.Abort()
			return
		}

		c.Next()
	}
}
