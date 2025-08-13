package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slmbngl/OrderAplication/internal/models"
	"github.com/slmbngl/OrderAplication/internal/repository"
	"github.com/slmbngl/OrderAplication/internal/service"
)

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserResponseReq true "User registration data"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Internal server error"
// @Router /api/auth/register [post]
func Register(c *fiber.Ctx) error {
	var req models.UserResponseReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Geçersiz giriş"})
	}

	hashedPassword := service.HashPassword(req.Password)

	user := &models.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	userRepo := repository.NewUserRepository()
	createdUser, err := userRepo.Create(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Registration failed: " + err.Error()})
	}

	createdUser.PasswordHash = ""

	return c.Status(201).JSON(createdUser)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserResponseReq true "User login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid input"
// @Failure 401 {string} string "Invalid username or password"
// @Failure 500 {string} string "Token could not be created"
// @Router /api/auth/login [post]
func Login(c *fiber.Ctx) error {
	var req models.UserResponseReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.GetByUsername(req.Username)
	if err != nil || !dbUser.IsActive {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid username or password"})
	}

	if service.HashPassword(req.Password) != dbUser.PasswordHash {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid username or password"})
	}

	token, err := service.GenerateJWT(dbUser.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token could not be created"})
	}

	return c.JSON(fiber.Map{
		"token":    token,
		"user_id":  dbUser.ID,
		"username": dbUser.Username,
	})
}
