package main

import (
	"log"

	"github.com/deepwrite/submission-service/internal/config"
	"github.com/deepwrite/submission-service/internal/database"
	"github.com/deepwrite/submission-service/internal/handler"
	"github.com/deepwrite/submission-service/internal/middleware"
	"github.com/deepwrite/submission-service/internal/repository"
	"github.com/deepwrite/submission-service/internal/routes"
	"github.com/deepwrite/submission-service/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	journalRepo := repository.NewJournalRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	journalService := service.NewJournalService(journalRepo)
	submissionService := service.NewSubmissionService(submissionRepo, journalRepo)
	reviewService := service.NewReviewService(reviewRepo, submissionRepo)

	journalHandler := handler.NewJournalHandler(journalService)
	submissionHandler := handler.NewSubmissionHandler(submissionService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	recommendHandler := handler.NewRecommendHandler(journalService, submissionService)
	formatCheckHandler := handler.NewFormatCheckHandler()

	jwtAuth := middleware.NewJWTAuth(&cfg.JWT)

	router := gin.Default()
	routes.Setup(router, journalHandler, submissionHandler, reviewHandler, recommendHandler, formatCheckHandler, jwtAuth)

	log.Printf("Submission service starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
