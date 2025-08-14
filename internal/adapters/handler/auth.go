package handler

import (
	"strconv"

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
		Role:         req.Role,
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

	token, err := service.GenerateJWT(dbUser.ID, dbUser.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token could not be created"})
	}

	return c.JSON(fiber.Map{
		"token":    token,
		"user_id":  dbUser.ID,
		"username": dbUser.Username,
		"role":     dbUser.Role,
	})
}

// GetAllUsers godoc
// @Summary Get all users (Admin only)
// @Description Get list of all users
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/users [get]
func GetAllUsers(c *fiber.Ctx) error {
	userRepo := repository.NewUserRepository()
	users, err := userRepo.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(users)
}

// UpdateUserRole godoc
// @Summary Update user role (Admin only)
// @Description Update a user's role
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param role body map[string]string true "Role data"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/users/{id}/role [put]
func UpdateUserRole(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var roleReq map[string]string
	if err := c.BodyParser(&roleReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
	}

	role, exists := roleReq["role"]
	if !exists {
		return c.Status(400).JSON(fiber.Map{"error": "Role is required"})
	}

	validRoles := map[string]bool{
		"user":  true,
		"admin": true,
	}

	if !validRoles[role] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role"})
	}

	userRepo := repository.NewUserRepository()

	_, err = userRepo.GetByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	err = userRepo.UpdateUserRole(userID, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User role updated successfully"})
}

// GetMe godoc
// @Summary Get current user profile
// @Description Get the profile information of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Router /api/auth/profile [get]
func GetMe(c *fiber.Ctx) error {
	// Assuming user ID is stored in locals by JWT middleware as "user_id"
	userIDVal := c.Locals("user_id")
	userID, ok := userIDVal.(int)
	if !ok {
		// Try to convert from string if middleware stores as string
		userIDStr, ok := userIDVal.(string)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		id, err := strconv.Atoi(userIDStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		userID = id
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.GetByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}
