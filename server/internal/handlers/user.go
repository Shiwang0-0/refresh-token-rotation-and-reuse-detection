package handlers

import (
	"errors"
	"fmt"

	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/dto"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/services"
	"github.com/Shiwang0-0/refresh-token-rotation-and-reuse-detection/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService services.UserService // depends on interface
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) UserRegister(c *fiber.Ctx) error {

	context := fiber.Map{
		"message": "user successfully registered",
	}

	var data dto.RegisterRequest

	if err := c.BodyParser(&data); err != nil {
		context["message"] = err.Error()
		return c.Status(400).JSON(context)
	}

	if validationErrors := utils.ValidateStruct(data); len(validationErrors) > 0 {
		return c.Status(422).JSON(fiber.Map{
			"message": "validation failed",
			"errors":  validationErrors,
		})
	}

	user, err := h.userService.RegisterUser(data)
	if err != nil {
		context["message"] = err.Error()
		if errors.Is(err, utils.ErrEmailTaken) {
			return c.Status(409).JSON(context)
		}
		return c.Status(400).JSON(context)
	}

	context["user"] = user

	return c.Status(200).JSON(context)
}

func (h *UserHandler) UserLogin(c *fiber.Ctx) error {

	context := fiber.Map{
		"message": "user successfully logged in",
	}

	var data dto.LoginRequest

	if err := c.BodyParser(&data); err != nil {
		context["message"] = err.Error()
		return c.Status(400).JSON(context)
	}

	if validationErrors := utils.ValidateStruct(data); len(validationErrors) > 0 {
		return c.Status(422).JSON(fiber.Map{
			"message": "validation failed",
			"errors":  validationErrors,
		})
	}

	user, accessToken, refreshToken, err := h.userService.LoginUser(data)

	if err != nil {
		if errors.Is(err, utils.ErrInvalidCredentials) {
			context["message"] = "invalid email or password"
			return c.Status(401).JSON(context)
		}
		context["message"] = err.Error()
		return c.Status(400).JSON(context)
	}

	// refresh token stored in http-only secure cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   7 * 24 * 60 * 60,
	})

	// access token sent in response, to be stored in react state
	context["access_token"] = accessToken
	context["user"] = user

	return c.Status(200).JSON(context)
}

func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	context := fiber.Map{
		"message": "user profile fetched",
	}

	userID := c.Locals("userID").(int)

	user, err := h.userService.GetUserProfile(userID)
	if err != nil {
		context["message"] = err.Error()
		return err
	}

	context["user"] = user
	return c.Status(200).JSON(context)
}

func (h *UserHandler) RefreshTokens(c *fiber.Ctx) error {
	context := fiber.Map{
		"message": "refreshed access token and refresh token",
	}

	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		fmt.Println("this")
		context["message"] = "authentication error, probably user is logged out"
		return c.Status(401).JSON(context)
	}

	newAccessToken, newRefreshToken, err := h.userService.RefreshTokens(refreshToken)
	if err != nil {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			MaxAge:   -1, // expire the cookie
		})
		context["message"] = "session expired, please login again"
		return c.Status(401).JSON(context)
	}

	fmt.Print("NEW TOKENS ON REFRESH: ", newRefreshToken, newAccessToken)

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   7 * 24 * 60 * 60,
	})

	// access token sent in response, to be stored in react state
	context["access_token"] = newAccessToken

	return c.Status(200).JSON(context)
}

func (h *UserHandler) UserLogout(c *fiber.Ctx) error {
	fmt.Println("logout hit")
	context := fiber.Map{
		"message": "user successfully logged out",
	}
	refreshToken := c.Cookies("refresh_token")

	if refreshToken != "" {
		// clear refresh token from DB
		_ = h.userService.LogoutUser(refreshToken)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   -1, // forces immediate expiry
	})

	return c.Status(200).JSON(context)
}
