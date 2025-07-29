package handlers

import (
	"fmt"
	"net/http"
	"time"
	"sjek/internal/database"
	"sjek/internal/middleware"
	"sjek/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Request struct untuk register tanpa field Type
type RegisterUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`    // Bisa username atau email
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string `json:"message" example:"operation successful"`
}

// Fungsi helper untuk register user
func registerUser(c *gin.Context, userType models.UserType) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah username sudah ada
	var existingUser models.User
	result := database.DB.Where("username = ?", req.Username).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
		return
	}

	// Cek apakah email sudah ada
	result = database.DB.Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Set activated date untuk user baru
	now := time.Now()

	// Create user dengan type yang sudah ditentukan
	user := models.User{
		ID:            uuid.New(),
		Username:      req.Username,
		Email:         req.Email,
		Password:      string(hashedPassword),
		Type:          userType, // Type otomatis sesuai endpoint
		Status:        models.UserStatusActive,
		ActivatedDate: &now,
	}

	// Save to database dalam transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create user: %v", err)})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{Token: token})
}

// @Summary      Register new admin
// @Description  Register a new admin user (Type automatically set to ADMIN)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterUserRequest true "Admin register credentials"
// @Success      201  {object}  AuthResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /register/admin [post]
func RegisterAdmin(c *gin.Context) {
	registerUser(c, models.UserTypeAdmin)
}

// @Summary      Register new driver
// @Description  Register a new driver user (Type automatically set to DRIVER)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterUserRequest true "Driver register credentials"
// @Success      201  {object}  AuthResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /register/driver [post]
func RegisterDriver(c *gin.Context) {
	registerUser(c, models.UserTypeDriver)
}

// @Summary      Login user
// @Description  Login with username/email and password to get JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials (username or email)"
// @Success      200  {object}  AuthResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log failed login attempt
		saveLoginLog(c, "", req.Login, "", models.LoginStatusFailed, "Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user dengan roles - cari berdasarkan username ATAU email
	var user models.User
	result := database.DB.Preload("Roles").Where("username = ? OR email = ?", req.Login, req.Login).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Log failed login attempt - user not found
			saveLoginLog(c, "", req.Login, "", models.LoginStatusFailed, "User not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			// Log failed login attempt - database error
			saveLoginLog(c, "", req.Login, "", models.LoginStatusFailed, "Database error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		}
		return
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		// Log failed login attempt - wrong password
		saveLoginLog(c, user.ID.String(), user.Username, user.Email, models.LoginStatusFailed, "Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get user roles
	var roleNames []string
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, roleNames)
	if err != nil {
		// Log failed login attempt - token generation error
		saveLoginLog(c, user.ID.String(), user.Username, user.Email, models.LoginStatusFailed, "Token generation failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Save token to database
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	if err := middleware.SaveTokenToDB(user.ID.String(), token, ipAddress, userAgent); err != nil {
		// Log failed login attempt - token save error
		saveLoginLog(c, user.ID.String(), user.Username, user.Email, models.LoginStatusFailed, "Failed to save token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	// Log successful login
	saveLoginLog(c, user.ID.String(), user.Username, user.Email, models.LoginStatusSuccess, "Login successful")

	c.JSON(http.StatusOK, AuthResponse{Token: token})
}

// Helper function untuk save login log
func saveLoginLog(c *gin.Context, userID, username, email, status, message string) {
	// Get IP address
	ipAddress := c.ClientIP()
	
	// Get User Agent
	userAgent := c.GetHeader("User-Agent")

	// Create login log
	loginLog := models.LoginLog{
		ID:        uuid.New(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Username:  username,
		Email:     email,
		Status:    status,
		Message:   message,
		LoginTime: time.Now(),
	}

	// Set UserID jika ada
	if userID != "" {
		if parsedUserID, err := uuid.Parse(userID); err == nil {
			loginLog.UserID = parsedUserID
		}
	}

	// Save to database (non-blocking, jangan sampai mengganggu login process)
	go func() {
		if err := database.DB.Create(&loginLog).Error; err != nil {
			fmt.Printf("Failed to save login log: %v\n", err)
		}
	}()
}
