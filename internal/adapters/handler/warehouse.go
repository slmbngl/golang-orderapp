package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/slmbngl/OrderAplication/internal/models"
	"github.com/slmbngl/OrderAplication/internal/repository"
)

var warehouseRepo = repository.NewWarehouseRepository()

// Warehouse Management Handlers

// @Summary Create warehouse
// @Description Create a new warehouse (Admin only)
// @Tags warehouses
// @Accept json
// @Produce json
// @Param warehouse body models.CreateWarehouseRequest true "Warehouse data"
// @Success 201 {object} models.Warehouse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses [post]
func CreateWarehouse(c *fiber.Ctx) error {
	var req models.CreateWarehouseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	warehouse, err := warehouseRepo.CreateWarehouse(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create warehouse",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(warehouse)
}

// @Summary Get all warehouses
// @Description Get list of all warehouses (Admin only)
// @Tags warehouses
// @Produce json
// @Success 200 {array} models.Warehouse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses [get]
func GetAllWarehouses(c *fiber.Ctx) error {
	warehouses, err := warehouseRepo.GetAllWarehouses()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get warehouses",
		})
	}

	return c.JSON(warehouses)
}

// @Summary Get warehouse by ID
// @Description Get warehouse details by ID (Admin only)
// @Tags warehouses
// @Produce json
// @Param id path int true "Warehouse ID"
// @Success 200 {object} models.Warehouse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{id} [get]
func GetWarehouseByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	warehouse, err := warehouseRepo.GetWarehouseByID(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Warehouse not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get warehouse",
		})
	}

	return c.JSON(warehouse)
}

// @Summary Update warehouse
// @Description Update warehouse details (Admin only)
// @Tags warehouses
// @Accept json
// @Produce json
// @Param id path int true "Warehouse ID"
// @Param warehouse body models.UpdateWarehouseRequest true "Warehouse data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{id} [put]
func UpdateWarehouse(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	var req models.UpdateWarehouseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = warehouseRepo.UpdateWarehouse(id, &req)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Warehouse not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update warehouse",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Warehouse updated successfully",
	})
}

// @Summary Delete warehouse
// @Description Delete warehouse (Admin only)
// @Tags warehouses
// @Produce json
// @Param id path int true "Warehouse ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{id} [delete]
func DeleteWarehouse(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	err = warehouseRepo.DeleteWarehouse(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Warehouse not found",
			})
		}
		// Check for custom error type
		if warehouseErr, ok := err.(*repository.WarehouseHasStockError); ok {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":        "Cannot delete warehouse with existing stock",
				"warehouse_id": warehouseErr.WarehouseID,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete warehouse",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Warehouse deleted successfully",
	})
}

// Stock Management Handlers

// @Summary Get warehouse stocks
// @Description Get all stocks for a specific warehouse (Admin only)
// @Tags warehouse-stocks
// @Produce json
// @Param id path int true "Warehouse ID"
// @Success 200 {array} models.WarehouseStock
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{id}/stocks [get]
func GetWarehouseStocks(c *fiber.Ctx) error {
	warehouseID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	stocks, err := warehouseRepo.GetWarehouseStocks(warehouseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get warehouse stocks",
		})
	}

	return c.JSON(stocks)
}

// @Summary Get product stock in warehouse
// @Description Get specific product stock in warehouse (Admin only)
// @Tags warehouse-stocks
// @Produce json
// @Param warehouseId path int true "Warehouse ID"
// @Param productId path int true "Product ID"
// @Success 200 {object} models.WarehouseStock
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{warehouseId}/stocks/{productId} [get]
func GetProductStockInWarehouse(c *fiber.Ctx) error {
	warehouseID, err := strconv.Atoi(c.Params("warehouseId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	productID, err := strconv.Atoi(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	stock, err := warehouseRepo.GetProductStockInWarehouse(warehouseID, productID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get stock",
		})
	}

	return c.JSON(stock)
}

// @Summary Update stock
// @Description Update product stock in warehouse (Admin only)
// @Tags warehouse-stocks
// @Accept json
// @Produce json
// @Param warehouseId path int true "Warehouse ID"
// @Param productId path int true "Product ID"
// @Param stock body models.UpdateStockRequest true "Stock data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{warehouseId}/stocks/{productId} [put]
func UpdateStock(c *fiber.Ctx) error {
	warehouseID, err := strconv.Atoi(c.Params("warehouseId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	productID, err := strconv.Atoi(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var req models.UpdateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = warehouseRepo.UpdateStock(warehouseID, productID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update stock",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stock updated successfully",
	})
}

// @Summary Add stock
// @Description Add stock to product in warehouse (Admin only)
// @Tags warehouse-stocks
// @Accept json
// @Produce json
// @Param warehouseId path int true "Warehouse ID"
// @Param productId path int true "Product ID"
// @Param stock body models.UpdateStockRequest true "Stock data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/warehouses/{warehouseId}/stocks/{productId}/add [post]
func AddStock(c *fiber.Ctx) error {
	warehouseID, err := strconv.Atoi(c.Params("warehouseId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid warehouse ID",
		})
	}

	productID, err := strconv.Atoi(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var req models.UpdateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = warehouseRepo.AddStock(warehouseID, productID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add stock",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stock added successfully",
	})
}

// @Summary Get all stocks
// @Description Get all stocks from all warehouses (Requires authentication)
// @Tags stocks
// @Produce json
// @Success 200 {array} models.WarehouseStock
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/stocks [get]
func GetAllStocks(c *fiber.Ctx) error {
	stocks, err := warehouseRepo.GetAllStocks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get stocks",
		})
	}

	return c.JSON(stocks)
}

// Transfer Management Handlers

// @Summary Create stock transfer
// @Description Create a new stock transfer (Requires authentication)
// @Tags transfers
// @Accept json
// @Produce json
// @Param transfer body models.StockTransferRequest true "Transfer data"
// @Success 201 {object} models.StockTransfer
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/transfers [post]
func CreateStockTransfer(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var req models.StockTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	transfer, err := warehouseRepo.CreateStockTransfer(&req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create transfer",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(transfer)
}

// @Summary Get all transfers
// @Description Get list of all stock transfers (Requires authentication)
// @Tags transfers
// @Produce json
// @Success 200 {array} models.StockTransfer
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/transfers [get]
func GetAllTransfers(c *fiber.Ctx) error {
	transfers, err := warehouseRepo.GetAllTransfers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get transfers",
		})
	}

	return c.JSON(transfers)
}

// @Summary Get transfer by ID
// @Description Get transfer details by ID (Requires authentication)
// @Tags transfers
// @Produce json
// @Param id path int true "Transfer ID"
// @Success 200 {object} models.StockTransfer
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/transfers/{id} [get]
func GetTransferByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transfer ID",
		})
	}

	transfer, err := warehouseRepo.GetTransferByID(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Transfer not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get transfer",
		})
	}

	return c.JSON(transfer)
}

// @Summary Update transfer status
// @Description Update transfer status (Admin only)
// @Tags transfers
// @Accept json
// @Produce json
// @Param id path int true "Transfer ID"
// @Param status body models.StockTransferStatusRequest true "Status data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/transfers/{id}/status [put]
func UpdateTransferStatus(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transfer ID",
		})
	}

	var req models.StockTransferStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate status
	validStatuses := []string{"pending", "completed", "failed", "cancelled"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Valid statuses: pending, completed, failed, cancelled",
		})
	}

	err = warehouseRepo.UpdateTransferStatus(id, req.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Transfer not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update transfer status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Transfer status updated successfully",
	})
}

// @Summary Process transfer
// @Description Process pending transfer (Admin only)
// @Tags transfers
// @Produce json
// @Param id path int true "Transfer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/transfers/{id}/process [post]
func ProcessTransfer(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transfer ID",
		})
	}

	err = warehouseRepo.ProcessTransfer(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Transfer not found",
			})
		}

		// Check for custom error types
		if transferErr, ok := err.(*repository.TransferNotPendingError); ok {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":          "Transfer is not in pending status",
				"current_status": transferErr.Status,
			})
		}

		if stockErr, ok := err.(*repository.InsufficientStockError); ok {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":        "Insufficient stock for transfer",
				"warehouse_id": stockErr.WarehouseID,
				"product_id":   stockErr.ProductID,
				"required":     stockErr.Required,
				"available":    stockErr.Available,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process transfer",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Transfer processed successfully",
	})
}
