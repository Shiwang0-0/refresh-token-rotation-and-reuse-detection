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

	user, accessToken, refreshToken, err := h.userService.RegisterUser(data)
	if err != nil {
		context["message"] = err.Error()
		if errors.Is(err, utils.ErrEmailTaken) {
			return c.Status(409).JSON(context)
		}
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

	return c.Status(201).JSON(context)
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
		context["message"] = "authentication error, probably user is logged out"
		return c.Status(401).JSON(context)
	}

	// as soon as you get the token in local variable, delete it from cookie, because new token will be issued now
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   -1, // expire the cookie
	})

	// get the user who holds this current refesh token
	// you dont validate from the cookie (because the token could still be live)
	// If the token is cryptographically valid but no longer exists in the DB, it may indicate refresh token reuse.
	user, err := h.userService.GetUserByRefreshToken(refreshToken)
	if err != nil {
		context["message"] = "cannot find user from refresh token"
		return c.Status(500).JSON(context)
	}

	// there does not exist any user with this refresh token
	// but it did exist in the cookie
	// token is valid but no longer exists in the DB.
	// This may indicate refresh token reuse.
	// valid user has the newly rotated token (because of token rotation), but hacker has the old one which no longer exist in the DB
	// therefore there is a possible chance valid user device is compromised
	// REUSE DETECTION
	if user == nil {
		// the hacker has the stolen token the token of the user
		// for the user, delete all the refresh tokens, we will logout from all sessions
		userID, err := utils.ValidateRefreshToken(refreshToken)

		// Expired token.
		// Periodic cleanup can remove expired session records from the database.
		if err != nil {
			fmt.Println("TOKEN WAS ALREADY EXPIRED... RUN A PERIODIC CLEANUP")
			// if error found, simply return authentication error, the session must have expired or invalid credentials
			context["message"] = "authentication error"
			return c.Status(401).JSON(context)
		}

		// if no error found, delete all sessions for the user
		fmt.Println("FOUND A SESSION THAT NO LONGER EXIST IN DB")
		fmt.Printf("REUSE DETECTION for user: %d\n", userID)
		fmt.Println("LOGGING OUT FROM ALL SESSIONS.....")
		// valid jwt but missing from DB
		// reuse-detection signal.
		if invalidateErr := h.userService.InvalidateAllUserSessions(userID); invalidateErr != nil {
			context["message"] = "error invalidating user sessions"
			return c.Status(500).JSON(context)
		}
		context["message"] = "refresh token reuse detected"
		return c.Status(401).JSON(context)
	}

	newAccessToken, newRefreshToken, err := h.userService.RefreshTokens(refreshToken)
	if err != nil {
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
