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

	// Admin endpoints (Admin role required)
	SetupAdminRoutes(api)

	// Warehouse management endpoints (Admin role required)
	SetupWarehouseRoutes(api)
}

func SetupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Get("/profile", middleware.JWTMiddleware(), handler.GetMe)
	auth.Post("/refresh", handler.RefreshToken)
	auth.Post("/logout", handler.Logout)
	auth.Post("/logout-all", middleware.JWTMiddleware(), handler.LogoutAllDevices)

}

func SetupProductRoutes(api fiber.Router) {
	products := api.Group("/products")
	products.Get("/", handler.GetProducts)
	products.Get("/:id", handler.GetProductByID)

	// Protected routes for product management
	products.Post("/", middleware.JWTMiddleware(), middleware.AdminMiddleware(), handler.CreateProduct)
	products.Put("/:id", middleware.JWTMiddleware(), middleware.AdminMiddleware(), handler.UpdateProduct)
	products.Delete("/:id", middleware.JWTMiddleware(), middleware.AdminMiddleware(), handler.DeleteProduct)
}

func SetupOrderRoutes(api fiber.Router) {
	orders := api.Group("/orders", middleware.JWTMiddleware())
	orders.Get("/", handler.GetOrders)
	orders.Get("/:id", handler.GetOrderByID)
	orders.Post("/", handler.CreateOrder)
	orders.Put("/:id/status", handler.UpdateOrderStatus)
	orders.Delete("/:id", handler.DeleteOrder)
}

func SetupAdminRoutes(api fiber.Router) {
	admin := api.Group("/admin", middleware.JWTMiddleware(), middleware.AdminMiddleware())
	admin.Get("/users", handler.GetAllUsers)             // List all users
	admin.Put("/users/:id/role", handler.UpdateUserRole) // Update user role
}

func SetupWarehouseRoutes(api fiber.Router) {
	// Warehouse management routes (JWT + Admin gerekli)
	warehouses := api.Group("/warehouses", middleware.JWTMiddleware(), middleware.AdminMiddleware())
	warehouses.Post("/", handler.CreateWarehouse)
	warehouses.Get("/", handler.GetAllWarehouses)
	warehouses.Get("/:id", handler.GetWarehouseByID)
	warehouses.Put("/:id", handler.UpdateWarehouse)
	warehouses.Delete("/:id", handler.DeleteWarehouse)

	// Warehouse-specific stock routes (JWT + Admin gerekli)
	warehouses.Get("/:id/stocks", handler.GetWarehouseStocks)
	warehouses.Get("/:warehouseId/stocks/:productId", handler.GetProductStockInWarehouse)
	warehouses.Put("/:warehouseId/stocks/:productId", handler.UpdateStock)
	warehouses.Post("/:warehouseId/stocks/:productId/add", handler.AddStock)

	// Global stock routes (Sadece JWT gerekli - görüntüleme için)
	stocks := api.Group("/stocks", middleware.JWTMiddleware())
	stocks.Get("/", handler.GetAllStocks)

	// Transfer management routes (Sadece JWT gerekli)
	transfers := api.Group("/transfers", middleware.JWTMiddleware())
	transfers.Post("/", handler.CreateStockTransfer)
	transfers.Get("/", handler.GetAllTransfers)
	transfers.Get("/:id", handler.GetTransferByID)

	// Admin only routes for transfer management (JWT + Admin gerekli)
	transfers.Put("/:id/status", middleware.AdminMiddleware(), handler.UpdateTransferStatus)
	transfers.Post("/:id/process", middleware.AdminMiddleware(), handler.ProcessTransfer)
}
