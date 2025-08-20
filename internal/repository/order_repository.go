package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/slmbngl/OrderAplication/internal/adapters/db"
	"github.com/slmbngl/OrderAplication/internal/models"
)

type OrderRepository interface {
	GetOrderByID(orderID, userID int) (*models.Order, error)
	GetOrderItems(orderID int) ([]models.OrderItem, error)
	CreateOrder(userID int, items []models.CreateOrderItemRequest) (*models.OrderWithItems, error)
	DeleteOrder(orderID, userID int) error
	UpdateOrderStatus(orderID, userID int, status string) error
}

type orderRepo struct{}

func NewOrderRepository() OrderRepository {
	return &orderRepo{}
}

func GetOrdersByUserID(userID int) ([]models.OrderWithItems, error) {
	// Önce siparişleri al
	orderRows, err := db.Pool.Query(context.Background(),
		`SELECT DISTINCT order_id, user_id, total_amount, status, created_at, username 
         FROM order_summary_view 
         WHERE user_id = $1 
         ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer orderRows.Close()

	var ordersWithItems []models.OrderWithItems
	for orderRows.Next() {
		var order models.Order
		err := orderRows.Scan(&order.ID, &order.UserID, &order.TotalAmount,
			&order.Status, &order.CreatedAt, &order.Username)
		if err != nil {
			return nil, err
		}

		// Her sipariş için items'ları al
		orderRepo := NewOrderRepository()
		items, err := orderRepo.GetOrderItems(order.ID)
		if err != nil {
			return nil, err
		}

		orderWithItems := models.OrderWithItems{
			Order: order,
			Items: items,
		}
		ordersWithItems = append(ordersWithItems, orderWithItems)
	}

	return ordersWithItems, nil
}
func (r *orderRepo) GetOrderByID(orderID, userID int) (*models.Order, error) {
	var order models.Order
	err := db.Pool.QueryRow(context.Background(),
		"SELECT id, user_id, total_amount, created_at FROM orders WHERE id = $1 AND user_id = $2",
		orderID, userID).Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.CreatedAt)

	if err != nil {
		return nil, err
	}

	order.Status = "pending"
	return &order, nil
}

func (r *orderRepo) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	itemRows, err := db.Pool.Query(context.Background(),
		`SELECT oi.id, oi.product_id, oi.quantity, p.name, p.description 
         FROM order_items oi 
         JOIN products p ON oi.product_id = p.id 
         WHERE oi.order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	var items []models.OrderItem
	for itemRows.Next() {
		var item models.OrderItem
		var productName, productDescription string
		err := itemRows.Scan(&item.ID, &item.ProductID, &item.Quantity, &productName, &productDescription)
		if err != nil {
			return nil, err
		}
		item.OrderID = orderID
		item.ProductName = productName
		item.ProductDescription = productDescription
		items = append(items, item)
	}

	return items, nil
}

func (r *orderRepo) CreateOrder(userID int, items []models.CreateOrderItemRequest) (*models.OrderWithItems, error) {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Create product repository instance
	productRepo := NewProductRepository()

	// Check warehouse stock for all items first
	for _, item := range items {
		_, err := productRepo.CheckWarehouseStock(item.ProductID, item.Quantity)
		if err != nil {
			if warehouseErr, ok := err.(*InsufficientWarehouseStockError); ok {
				return nil, errors.New("insufficient warehouse stock for product ID: " +
					strconv.Itoa(warehouseErr.ProductID))
			}
			return nil, err
		}

		// Additional check: ensure product exists and get details
		var productPrice float64
		var productName, productDescription string
		err = tx.QueryRow(context.Background(),
			"SELECT price, name, description FROM products WHERE id = $1",
			item.ProductID).Scan(&productPrice, &productName, &productDescription)
		if err != nil {
			return nil, err
		}
	}

	// Create order with total_amount = 0 initially
	var orderID int
	err = tx.QueryRow(context.Background(),
		"INSERT INTO orders (user_id, total_amount) VALUES ($1, $2) RETURNING id",
		userID, 0).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	// Add order items and calculate total
	var orderItems []models.OrderItem
	var totalAmount float64
	for _, item := range items {
		// Get product details again
		var productPrice float64
		var productName, productDescription string
		err = tx.QueryRow(context.Background(),
			"SELECT price, name, description FROM products WHERE id = $1",
			item.ProductID).Scan(&productPrice, &productName, &productDescription)
		if err != nil {
			return nil, err
		}

		// Insert order item
		var itemID int
		err = tx.QueryRow(context.Background(),
			"INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id",
			orderID, item.ProductID, item.Quantity).Scan(&itemID)
		if err != nil {
			return nil, err
		}

		orderItem := models.OrderItem{
			ID:                 itemID,
			OrderID:            orderID,
			ProductID:          item.ProductID,
			Quantity:           item.Quantity,
			Price:              productPrice,
			ProductName:        productName,
			ProductDescription: productDescription,
		}
		orderItems = append(orderItems, orderItem)
		totalAmount += float64(item.Quantity) * productPrice
	}

	// Update order with calculated total amount
	_, err = tx.Exec(context.Background(),
		"UPDATE orders SET total_amount = $1 WHERE id = $2",
		totalAmount, orderID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	// Return created order with items
	order := models.Order{
		ID:          orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      "pending",
	}

	orderWithItems := &models.OrderWithItems{
		Order: order,
		Items: orderItems,
	}

	return orderWithItems, nil
}

func (r *orderRepo) UpdateOrderStatus(orderID, userID int, status string) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get current order status first
	var currentStatus string
	err = tx.QueryRow(context.Background(),
		"SELECT status FROM orders WHERE id = $1 AND user_id = $2",
		orderID, userID).Scan(&currentStatus)
	if err != nil {
		return err
	}

	// If status is the same, no need to update
	if currentStatus == status {
		return nil
	}

	// Update order status
	result, err := tx.Exec(context.Background(),
		"UPDATE orders SET status = $1 WHERE id = $2 AND user_id = $3",
		status, orderID, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	// Create product repository instance for warehouse operations
	productRepo := NewProductRepository()

	// If status is confirmed, check and update warehouse stock
	if status == "confirmed" && currentStatus != "confirmed" {
		// Get order items
		items, err := r.GetOrderItems(orderID)
		if err != nil {
			return err
		}

		// Check and update warehouse stock for each item
		for _, item := range items {
			// Check warehouse stock availability
			_, err := productRepo.CheckWarehouseStock(item.ProductID, item.Quantity)
			if err != nil {
				if warehouseErr, ok := err.(*InsufficientWarehouseStockError); ok {
					return errors.New("insufficient warehouse stock for product: " + item.ProductName +
						", required: " + strconv.Itoa(warehouseErr.RequiredStock) +
						", available: " + strconv.Itoa(warehouseErr.AvailableStock))
				}
				return err
			}

			// Update warehouse stock (decrease)
			err = productRepo.UpdateWarehouseStock(item.ProductID, item.Quantity, "decrease")
			if err != nil {
				return err
			}
		}
	}

	// If status is cancelled or pending from confirmed, restore warehouse stock
	if (status == "cancelled" && currentStatus == "confirmed") ||
		(status == "pending" && currentStatus == "confirmed") {
		// Get order items
		items, err := r.GetOrderItems(orderID)
		if err != nil {
			return err
		}

		// Restore warehouse stock for each item
		for _, item := range items {
			// Update warehouse stock (increase)
			err = productRepo.UpdateWarehouseStock(item.ProductID, item.Quantity, "increase")
			if err != nil {
				return err
			}
		}
	}

	// Commit transaction
	return tx.Commit(context.Background())
}

func (r *orderRepo) DeleteOrder(orderID, userID int) error {
	result, err := db.Pool.Exec(context.Background(),
		"DELETE FROM orders WHERE id = $1 AND user_id = $2",
		orderID, userID)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
