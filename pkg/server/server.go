package server

import (
	"fmt"
	"log"

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
	fmt.Println("✓ postgres connected")

	// Init Redis
	rdb, err := database.InitRedis(cfg.App.Db.Redis)
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	fmt.Println("✓ redis connected")

	// Gin engine
	r := gin.Default()

	return &Server{
		Config: cfg,
		DB:     db,
		Redis:  rdb,
		Router: r,
	}
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
