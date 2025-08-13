package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/slmbngl/OrderAplication/internal/models"
	"github.com/slmbngl/OrderAplication/internal/repository"
)

// GetProducts godoc
// @Summary Get all products
// @Description Get all available products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {string} string "Internal server error"
// @Router /api/products [get]
func GetProducts(c *fiber.Ctx) error {
	productRepo := repository.NewProductRepository()
	products, err := productRepo.GetAllProducts()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(products)
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Get a specific product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {string} string "Product not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/products/{id} [get]
func GetProductByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	productRepo := repository.NewProductRepository()
	product, err := productRepo.GetProductByID(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(product)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product (Admin only)
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.ProductRequest true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/products [post]
func CreateProduct(c *fiber.Ctx) error {
	var productReq models.ProductRequest
	if err := c.BodyParser(&productReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
	}

	productRepo := repository.NewProductRepository()
	product, err := productRepo.CreateProduct(&productReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(product)
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product (Admin only)
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product body models.ProductRequest true "Product data"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/products/{id} [put]
func UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	var productReq models.ProductRequest
	if err := c.BodyParser(&productReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
	}

	productRepo := repository.NewProductRepository()
	err = productRepo.UpdateProduct(id, &productReq)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Product successfully updated"})
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product (Admin only)
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/products/{id} [delete]
func DeleteProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	productRepo := repository.NewProductRepository()
	err = productRepo.DeleteProduct(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Product successfully deleted"})
}
