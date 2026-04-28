package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dimas292/url_shortener/pkg/auth"
	"github.com/dimas292/url_shortener/pkg/config"
	"github.com/dimas292/url_shortener/pkg/database"
	"github.com/dimas292/url_shortener/pkg/router"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Server holds all shared dependencies and the Gin engine.
type Server struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	JWT    *auth.JWTService
	Router *gin.Engine
}

// New initializes the server: loads config, connects databases, sets up the router.
func New(configPath string) *Server {
	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Init Postgres
	db, err := database.InitPostgres(cfg.App.Db.Postgres)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	fmt.Println("postgres connected")

	// Init Redis
	rdb, err := database.InitRedis(cfg.App.Db.Redis)
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	fmt.Println("redis connected")

	// Init JWT
	jwtService := auth.NewJWTService(cfg.App.Jwt)
	fmt.Println("jwt initialized")

	// Gin engine
	r := gin.Default()

	srv := &Server{
		Config: cfg,
		DB:     db,
		Redis:  rdb,
		JWT:    jwtService,
		Router: r,
	}

	// Register health check endpoint
	srv.registerHealthCheck()

	return srv
}

// registerHealthCheck registers the GET /health endpoint.
func (s *Server) registerHealthCheck() {
	s.Router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})
}

// RegisterModules registers feature modules under /api/v1.
func (s *Server) RegisterModules(modules ...router.Module) {
	router.RegisterModules(s.Router, "/api/v1", modules...)
}

// Run starts the HTTP server on the configured port.
func (s *Server) Run() {
	port := s.Config.App.Port
	fmt.Printf("server running on %s\n", port)
	if err := s.Router.Run(port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
