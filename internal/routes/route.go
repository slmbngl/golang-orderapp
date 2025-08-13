package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/slmbngl/OrderAplication/internal/adapters/handler"
	"github.com/slmbngl/OrderAplication/internal/adapters/middleware"
)

func SetupRoutes(app *fiber.App) {
	// Swagger endpoint
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API groups
	api := app.Group("/api")
	// Auth endpoints (JWT not required)
	SetupAuthRoutes(api)

	// Product endpoints
	SetupProductRoutes(api)

	// Order endpoints (JWT required)
	SetupOrderRoutes(api)

}
func SetupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
}
func SetupProductRoutes(api fiber.Router) {
	products := api.Group("/products")
	products.Get("/", handler.GetProducts)
	products.Get("/:id", handler.GetProductByID)

	// Protected routes for product management
	products.Post("/", middleware.JWTMiddleware(), handler.CreateProduct)
	products.Put("/:id", middleware.JWTMiddleware(), handler.UpdateProduct)
	products.Delete("/:id", middleware.JWTMiddleware(), handler.DeleteProduct)
}
func SetupOrderRoutes(api fiber.Router) {
	orders := api.Group("/orders", middleware.JWTMiddleware())
	orders.Get("/", handler.GetOrders)
	orders.Get("/:id", handler.GetOrderByID)
	orders.Post("/", handler.CreateOrder)
	orders.Put("/:id/status", handler.UpdateOrderStatus)
	orders.Delete("/:id", handler.DeleteOrder)
}
