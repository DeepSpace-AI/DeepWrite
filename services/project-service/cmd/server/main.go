package main

import (
	"log"

	"github.com/deepwrite/project-service/internal/config"
	"github.com/deepwrite/project-service/internal/database"
	"github.com/deepwrite/project-service/internal/handler"
	"github.com/deepwrite/project-service/internal/middleware"
	"github.com/deepwrite/project-service/internal/repository"
	"github.com/deepwrite/project-service/internal/routes"
	"github.com/deepwrite/project-service/internal/service"
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

	projectRepo := repository.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)
	jwtAuth := middleware.NewJWTAuth(&cfg.JWT)

	router := gin.Default()
	routes.Setup(router, projectHandler, jwtAuth)

	log.Printf("Project service starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
