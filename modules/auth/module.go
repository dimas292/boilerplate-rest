package auth

import (
	pkgauth "github.com/dimas292/boilerplate-rest/pkg/auth"
	"github.com/dimas292/boilerplate-rest/pkg/response"
	"github.com/dimas292/boilerplate-rest/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AuthModule implements router.Module for authentication.
type AuthModule struct {
	service    *AuthService
	jwtService *pkgauth.JWTService
}

// NewAuthModule wires up the auth module's dependencies.
func NewAuthModule(db *gorm.DB, rdb *redis.Client, jwtService *pkgauth.JWTService) *AuthModule {
	// Auto-migrate
	db.AutoMigrate(&User{})

	service := NewAuthService(db, rdb, jwtService)

	return &AuthModule{
		service:    service,
		jwtService: jwtService,
	}
}

// JWTService returns the JWT service for use in other modules' middleware.
func (m *AuthModule) JWTService() *pkgauth.JWTService {
	return m.jwtService
}

// RegisterRoutes registers auth routes.
//
// Public:
//
//	POST /auth/register  → Register
//	POST /auth/login     → Login
//
// Protected (requires JWT):
//
//	GET  /auth/profile   → Get current user profile
func (m *AuthModule) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		// Public routes
		authGroup.POST("/register", m.handleRegister)
		authGroup.POST("/login", m.handleLogin)

		// Protected routes
		protected := authGroup.Group("")
		protected.Use(pkgauth.AuthMiddleware(m.jwtService))
		{
			protected.GET("/profile", m.handleProfile)
		}
	}
}

// handleRegister handles POST /auth/register.
func (m *AuthModule) handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	// Sanitize input
	validator.SanitizeStruct(&req)

	result, err := m.service.Register(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, "registered successfully", result)
}

// handleLogin handles POST /auth/login.
func (m *AuthModule) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	// Sanitize input
	validator.SanitizeStruct(&req)

	result, err := m.service.Login(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, "login successful", result)
}

// handleProfile handles GET /auth/profile (protected).
func (m *AuthModule) handleProfile(c *gin.Context) {
	userID := pkgauth.GetUserID(c)

	profile, err := m.service.GetProfile(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, "profile retrieved", profile)
}
