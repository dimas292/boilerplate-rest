package auth

import (
	"context"
	"errors"
	"time"

	"github.com/dimas292/boilerplate-rest/pkg/apperror"
	pkgauth "github.com/dimas292/boilerplate-rest/pkg/auth"
	"github.com/dimas292/boilerplate-rest/pkg/logger"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = apperror.Conflict("email already registered")
	ErrInvalidCredentials = apperror.Unauthorized("invalid email or password")
	ErrUserNotFound       = apperror.NotFound("user not found")
)

// AuthService handles authentication business logic.
type AuthService struct {
	db  *gorm.DB
	rdb *redis.Client
	jwt *pkgauth.JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *gorm.DB, rdb *redis.Client, jwt *pkgauth.JWTService) *AuthService {
	return &AuthService{db: db, rdb: rdb, jwt: jwt}
}

// Register creates a new user with hashed password.
func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if email already exists
	var count int64
	s.db.Model(&User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal("failed to process registration", err)
	}

	user := User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, apperror.Internal("failed to create user", err)
	}

	// Generate JWT
	token, err := s.jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, apperror.Internal("failed to generate token", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// Login authenticates a user and returns a JWT.
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	var user User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, apperror.Internal("failed to authenticate", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT
	token, err := s.jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, apperror.Internal("failed to generate token", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// GetProfile retrieves the current user's profile.
func (s *AuthService) GetProfile(userID string) (*UserResponse, error) {
	prefixUser := "cache:user:"
	keyUser := prefixUser + userID

	// Check Redis cache first (stored as Hash)
	cachedData, err := s.rdb.HGetAll(context.Background(), keyUser).Result()
	if err == nil && len(cachedData) > 0 {
		// Cache HIT
		return &UserResponse{
			ID:    cachedData["id"],
			Name:  cachedData["name"],
			Email: cachedData["email"],
			Role:  cachedData["role"],
		}, nil
	}

	if err != nil && err != redis.Nil {
		// Redis error (down/timeout) — log but continue to DB
		logger.Warn().Err(err).Str("key", keyUser).Msg("redis get profile failed, falling back to DB")
	}

	// Cache MISS — query DB
	var user User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Store in Redis cache
	userMap := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}

	if err := s.rdb.HSet(context.Background(), keyUser, userMap).Err(); err != nil {
		logger.Warn().Err(err).Str("key", keyUser).Msg("redis set profile failed")
	} else {
		s.rdb.Expire(context.Background(), keyUser, 30*time.Minute)
	}

	// Return data from DB
	resp := user.ToResponse()
	return &resp, nil
}