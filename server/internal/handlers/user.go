package handlers

import (
	"errors"

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

	accessToken, refreshToken, err := h.userService.LoginUser(data)

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
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   7 * 24 * 60 * 60,
	})

	// access token sent in response, to be stored in react state
	context["access_token"] = accessToken

	return c.Status(200).JSON(context)
}
