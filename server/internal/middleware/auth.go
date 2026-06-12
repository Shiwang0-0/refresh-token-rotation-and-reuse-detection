package middleware

import (
	"strings"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// for routes that requires access token
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"message": "unauthorized",
			})
		}

		// strip "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(401).JSON(fiber.Map{
				"message": "invalid token format",
			})
		}

		userID, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			// this is the 401 that api.js catches in fronted to call the refresh end point
			return c.Status(401).JSON(fiber.Map{
				"message": "token_expired",
			})
		}

		// got the userId from the accessToken, make it available to other routes
		c.Locals("userID", userID)
		return c.Next()
	}
}
