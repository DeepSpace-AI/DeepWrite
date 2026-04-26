package main

import (
	"log"

	"github.com/deepwrite/user-service/internal/config"
	"github.com/deepwrite/user-service/internal/database"
	"github.com/deepwrite/user-service/internal/handler"
	"github.com/deepwrite/user-service/internal/repository"
	"github.com/deepwrite/user-service/internal/routes"
	"github.com/deepwrite/user-service/internal/service"
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

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	userService := service.NewUserService(userRepo, &cfg.JWT)
	teamService := service.NewTeamService(teamRepo, userRepo)

	userHandler := handler.NewUserHandler(userService)
	teamHandler := handler.NewTeamHandler(teamService)

	router := gin.Default()
	routes.Setup(router, userHandler, teamHandler, userService)

	log.Printf("User service starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
