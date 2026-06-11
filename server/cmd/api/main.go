package main

import (
	"log"
	"os"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/config"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/repository"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/router"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	port := os.Getenv("PORT")

	db, err := config.LoadDB()
	if err != nil {
		log.Fatalf("MySQL connection failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()

	app.Use(logger.New())

	app.Use(config.CorsConfig)

	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	router.RouteSetup(app, userService)

	log.Fatal(app.Listen(":" + port))
}
