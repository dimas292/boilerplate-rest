package auth

import (
	"errors"
	"fmt"

	pkgauth "github.com/dimas292/url_shortener/pkg/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

// AuthService handles authentication business logic.
type AuthService struct {
	db  *gorm.DB
	jwt *pkgauth.JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *gorm.DB, jwt *pkgauth.JWTService) *AuthService {
	return &AuthService{db: db, jwt: jwt}
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
		return nil, fmt.Errorf("auth register hash: %w", err)
	}

	user := User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("auth register create: %w", err)
	}

	// Generate JWT
	token, err := s.jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("auth login query: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT
	token, err := s.jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// GetProfile retrieves the current user's profile.
func (s *AuthService) GetProfile(userID string) (*UserResponse, error) {
	var user User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	resp := user.ToResponse()
	return &resp, nil
}
