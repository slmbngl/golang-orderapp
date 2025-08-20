package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/slmbngl/OrderAplication/internal/models"
	"github.com/slmbngl/OrderAplication/internal/repository"
)

// GetOrders godoc
// @Summary Get user's orders
// @Description Get all orders for authenticated user
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Order
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/orders [get]
func GetOrders(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	orders, err := repository.GetOrdersByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(orders)
}

// GetOrderByID godoc
// @Summary Get order by ID
// @Description Get a specific order for authenticated user with items
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} models.OrderWithItems
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Order not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/orders/{id} [get]
func GetOrderByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	orderIDStr := c.Params("id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	orderRepo := repository.NewOrderRepository()

	order, err := orderRepo.GetOrderByID(orderID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Order not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get order items
	items, err := orderRepo.GetOrderItems(order.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	orderWithItems := models.OrderWithItems{
		Order: *order,
		Items: items,
	}

	return c.JSON(orderWithItems)
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for authenticated user
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.CreateOrderRequest true "Order data"
// @Success 201 {object} models.OrderWithItems
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/orders [post]
func CreateOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var orderReq models.CreateOrderRequest
	if err := c.BodyParser(&orderReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
	}

	if len(orderReq.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Order must contain at least one item"})
	}

	orderRepo := repository.NewOrderRepository()
	orderWithItems, err := orderRepo.CreateOrder(userID, orderReq.Items)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "Product not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(orderWithItems)
}

// DeleteOrder godoc
// @Summary Delete order
// @Description Delete an order for authenticated user
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/orders/{id} [delete]
func DeleteOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	orderIDStr := c.Params("id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	orderRepo := repository.NewOrderRepository()
	err = orderRepo.DeleteOrder(orderID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Order not found or you don't have permission to delete it"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Order successfully deleted"})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order, you can choose from "pending", "confirmed" or "cancelled"
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param status body models.UpdateOrderStatusRequest true "Status data"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/orders/{id}/status [put]
func UpdateOrderStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	orderIDStr := c.Params("id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	var statusReq map[string]string
	if err := c.BodyParser(&statusReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
	}

	status, exists := statusReq["status"]
	if !exists {
		return c.Status(400).JSON(fiber.Map{"error": "Status is required"})
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"cancelled": true,
	}

	if !validStatuses[status] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid status"})
	}

	orderRepo := repository.NewOrderRepository()
	err = orderRepo.UpdateOrderStatus(orderID, userID, status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Order not found or you don't have permission to update it"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Order status successfully updated"})
}
