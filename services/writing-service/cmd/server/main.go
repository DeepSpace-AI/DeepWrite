package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepwrite/writing-service/internal/config"
	"github.com/deepwrite/writing-service/internal/handler"
	"github.com/deepwrite/writing-service/internal/middleware"
	"github.com/deepwrite/writing-service/internal/repository"
	"github.com/deepwrite/writing-service/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.PostgresURL())
	if err != nil {
		log.Fatalf("Failed to open PostgreSQL connection: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURL()))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		mongoClient.Disconnect(ctx)
	}()

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	mongoDB := mongoClient.Database(cfg.MongoDB)

	// Initialize repositories
	docRepo := repository.NewPostgresDocumentRepository(db)
	contentRepo := repository.NewMongoContentRepository(mongoDB)
	versionRepo := repository.NewPostgresVersionRepository(db)
	snapshotRepo := repository.NewMongoSnapshotRepository(mongoDB)
	collabRepo := repository.NewPostgresCollaboratorRepository(db)

	// Initialize services
	docService := service.NewDocumentService(docRepo, contentRepo, versionRepo, snapshotRepo, collabRepo)
	collabHub := service.NewCollaborationHub(contentRepo)
	go collabHub.Run()

	// Initialize handlers
	docHandler := handler.NewDocumentHandler(docService)
	wsHandler := handler.NewWebSocketHandler(collabHub, cfg.JWTSecret)

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api/documents")
	api.Use(middleware.Auth(cfg.JWTSecret))
	{
		api.POST("", docHandler.CreateDocument)
		api.GET("", docHandler.ListDocuments)
		api.GET("/:id", docHandler.GetDocument)
		api.PUT("/:id", docHandler.UpdateDocument)
		api.PUT("/:id/content", docHandler.UpdateContent)
		api.DELETE("/:id", docHandler.DeleteDocument)

		api.POST("/:id/versions", docHandler.CreateVersion)
		api.GET("/:id/versions", docHandler.ListVersions)

		api.POST("/:id/collaborators", docHandler.AddCollaborator)
		api.GET("/:id/collaborators", docHandler.GetCollaborators)
		api.DELETE("/:id/collaborators/:uid", docHandler.RemoveCollaborator)

		api.GET("/:id/ws", wsHandler.HandleWebSocket)
		api.GET("/:id/active-users", wsHandler.GetActiveUsers)
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Writing service starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
