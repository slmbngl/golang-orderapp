// @title           Order Management API
// @version         1.0
// @description     This is an order management API with user authentication, product management, and order processing
// @host            localhost:4504
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description please enter your token with "Bearer " prefix for JWT token

package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/slmbngl/OrderAplication/docs" // Swagger docs
	"github.com/slmbngl/OrderAplication/internal/adapters/db"
	"github.com/slmbngl/OrderAplication/internal/routes"
)

func main() {
	// Database connection
	db.Connect()

	// Initialize Fiber app
	app := fiber.New()

	// Middlewares
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	routes.SetupRoutes(app)
	// START SERVER
	log.Fatal(app.Listen(":4504"))
}
