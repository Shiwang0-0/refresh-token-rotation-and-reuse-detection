package router

import (
	"fmt"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/handlers"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/services"
	"github.com/gofiber/fiber/v2"
)

func RouteSetup(app *fiber.App, userService services.UserService) {
	fmt.Println("Registering routes")
	userHandler := handlers.NewUserHandler(userService)

	api := app.Group("/api")
	api.Post("/login", userHandler.UserLogin)
	api.Post("/register", userHandler.UserRegister)

}
