package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/slmbngl/OrderAplication/internal/adapters/db"
	"github.com/slmbngl/OrderAplication/internal/models"
)

type ProductRepository interface {
	GetAllProducts() ([]models.Product, error)
	GetProductByID(id int) (*models.Product, error)
	CreateProduct(productReq *models.ProductRequest) (*models.Product, error)
	UpdateProduct(id int, productReq *models.ProductRequest) error
	DeleteProduct(id int) error
	CheckWarehouseStock(productID, quantity int) (*models.WarehouseStock, error)
	UpdateWarehouseStock(productID, quantity int, operation string) error
}

type productRepo struct{}

func NewProductRepository() ProductRepository {
	return &productRepo{}
}

func (r *productRepo) GetAllProducts() ([]models.Product, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT p.id, p.name, p.description, p.price, p.stock, p.warehouse_id, p.created_at, w.name 
         FROM products p 
         JOIN warehouses w ON p.warehouse_id = w.id 
         ORDER BY p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.WarehouseID, &p.CreatedAt, &p.WarehouseName)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *productRepo) GetProductByID(id int) (*models.Product, error) {
	var p models.Product
	err := db.Pool.QueryRow(context.Background(),
		`SELECT p.id, p.name, p.description, p.price, p.stock, p.warehouse_id, p.created_at, w.name
         FROM products p 
         JOIN warehouses w ON p.warehouse_id = w.id 
         WHERE p.id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.WarehouseID, &p.CreatedAt, &p.WarehouseName)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *productRepo) CreateProduct(productReq *models.ProductRequest) (*models.Product, error) {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Create product
	var product models.Product
	err = tx.QueryRow(context.Background(),
		`INSERT INTO products (name, description, price, stock, warehouse_id) 
         VALUES ($1, $2, $3, $4, $5) 
         RETURNING id, name, description, price, stock, warehouse_id, created_at`,
		productReq.Name, productReq.Description, productReq.Price,
		productReq.Stock, productReq.WarehouseID).
		Scan(&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.WarehouseID, &product.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Create or update warehouse stock
	_, err = tx.Exec(context.Background(),
		`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity) 
         VALUES ($1, $2, $3)
         ON CONFLICT (warehouse_id, product_id) 
         DO UPDATE SET quantity = warehouse_stocks.quantity + $3, updated_at = CURRENT_TIMESTAMP`,
		productReq.WarehouseID, product.ID, productReq.Stock)

	if err != nil {
		return nil, err
	}

	// Commit transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepo) UpdateProduct(id int, productReq *models.ProductRequest) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get current product info
	var currentWarehouseID, currentStock int
	err = tx.QueryRow(context.Background(),
		"SELECT warehouse_id, stock FROM products WHERE id = $1", id).
		Scan(&currentWarehouseID, &currentStock)
	if err != nil {
		return err
	}

	// Update product
	result, err := tx.Exec(context.Background(),
		"UPDATE products SET name=$1, description=$2, price=$3, stock=$4, warehouse_id=$5 WHERE id=$6",
		productReq.Name, productReq.Description, productReq.Price,
		productReq.Stock, productReq.WarehouseID, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	// Update warehouse stocks if warehouse changed
	if currentWarehouseID != productReq.WarehouseID {
		// Remove from old warehouse
		_, err = tx.Exec(context.Background(),
			`UPDATE warehouse_stocks SET quantity = quantity - $1, updated_at = CURRENT_TIMESTAMP
             WHERE warehouse_id = $2 AND product_id = $3`,
			currentStock, currentWarehouseID, id)
		if err != nil {
			return err
		}

		// Add to new warehouse
		_, err = tx.Exec(context.Background(),
			`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity) 
             VALUES ($1, $2, $3)
             ON CONFLICT (warehouse_id, product_id) 
             DO UPDATE SET quantity = warehouse_stocks.quantity + $3, updated_at = CURRENT_TIMESTAMP`,
			productReq.WarehouseID, id, productReq.Stock)
		if err != nil {
			return err
		}
	} else {
		// Same warehouse, update stock difference
		stockDiff := productReq.Stock - currentStock
		if stockDiff != 0 {
			_, err = tx.Exec(context.Background(),
				`UPDATE warehouse_stocks SET quantity = quantity + $1, updated_at = CURRENT_TIMESTAMP
                 WHERE warehouse_id = $2 AND product_id = $3`,
				stockDiff, productReq.WarehouseID, id)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(context.Background())
}

func (r *productRepo) DeleteProduct(id int) error {
	// Begin transaction for cascading delete
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// First, get all order IDs that contain this product
	orderRows, err := tx.Query(context.Background(),
		"SELECT DISTINCT order_id FROM order_items WHERE product_id = $1", id)
	if err != nil {
		return err
	}
	defer orderRows.Close()

	var orderIDs []int
	for orderRows.Next() {
		var orderID int
		err := orderRows.Scan(&orderID)
		if err != nil {
			return err
		}
		orderIDs = append(orderIDs, orderID)
	}

	// Delete order items for this product
	_, err = tx.Exec(context.Background(),
		"DELETE FROM order_items WHERE product_id = $1", id)
	if err != nil {
		return err
	}

	// Delete orders that now have no items left
	for _, orderID := range orderIDs {
		var remainingItems int
		err = tx.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM order_items WHERE order_id = $1", orderID).Scan(&remainingItems)
		if err != nil {
			return err
		}

		// If no items left in the order, delete the order
		if remainingItems == 0 {
			_, err = tx.Exec(context.Background(),
				"DELETE FROM orders WHERE id = $1", orderID)
			if err != nil {
				return err
			}
		}
	}

	// Delete warehouse stocks
	_, err = tx.Exec(context.Background(),
		"DELETE FROM warehouse_stocks WHERE product_id = $1", id)
	if err != nil {
		return err
	}

	// Finally, delete the product
	result, err := tx.Exec(context.Background(),
		"DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	// Commit the transaction
	return tx.Commit(context.Background())
}

func (r *productRepo) CheckWarehouseStock(productID, quantity int) (*models.WarehouseStock, error) {
	// Önce ürünün hangi warehouse'larda stoku olduğunu kontrol et
	rows, err := db.Pool.Query(context.Background(),
		`SELECT ws.id, ws.warehouse_id, ws.product_id, ws.quantity, ws.reserved_quantity,
                ws.created_at, ws.updated_at, w.name, p.name, p.price
         FROM warehouse_stocks ws
         JOIN warehouses w ON ws.warehouse_id = w.id
         JOIN products p ON ws.product_id = p.id
         WHERE ws.product_id = $1 AND ws.quantity >= $2
         ORDER BY ws.quantity DESC
         LIMIT 1`,
		productID, quantity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		// Hiçbir warehouse'da yeterli stok yok, tüm stokları kontrol et
		var totalStock, totalAvailable int
		err = db.Pool.QueryRow(context.Background(),
			`SELECT COALESCE(SUM(quantity), 0), COALESCE(SUM(quantity - reserved_quantity), 0)
             FROM warehouse_stocks WHERE product_id = $1`,
			productID).Scan(&totalStock, &totalAvailable)
		if err != nil {
			return nil, err
		}

		return nil, &InsufficientWarehouseStockError{
			ProductID:      productID,
			WarehouseID:    0, // No specific warehouse
			RequiredStock:  quantity,
			AvailableStock: totalAvailable,
		}
	}

	var stock models.WarehouseStock
	err = rows.Scan(&stock.ID, &stock.WarehouseID, &stock.ProductID, &stock.Quantity,
		&stock.ReservedQuantity, &stock.CreatedAt, &stock.UpdatedAt,
		&stock.WarehouseName, &stock.ProductName, &stock.ProductPrice)
	if err != nil {
		return nil, err
	}

	stock.AvailableStock = stock.Quantity - stock.ReservedQuantity

	// Check if available stock is sufficient
	if stock.AvailableStock < quantity {
		return nil, &InsufficientWarehouseStockError{
			ProductID:      productID,
			WarehouseID:    stock.WarehouseID,
			RequiredStock:  quantity,
			AvailableStock: stock.AvailableStock,
		}
	}

	return &stock, nil
}

func (r *productRepo) UpdateWarehouseStock(productID, quantity int, operation string) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	var updateQuery string
	var productUpdateQuery string

	switch operation {
	case "decrease":
		// Önce hangi warehouse'dan düşeceğimizi belirle (en fazla stoku olan)
		var warehouseID int
		err = tx.QueryRow(context.Background(),
			`SELECT warehouse_id FROM warehouse_stocks 
             WHERE product_id = $1 AND quantity >= $2 
             ORDER BY quantity DESC LIMIT 1`,
			productID, quantity).Scan(&warehouseID)
		if err != nil {
			return err
		}

		updateQuery = `UPDATE warehouse_stocks SET quantity = quantity - $1, updated_at = CURRENT_TIMESTAMP 
                       WHERE product_id = $2 AND warehouse_id = $3`
		_, err = tx.Exec(context.Background(), updateQuery, quantity, productID, warehouseID)

		// Products tablosunu da güncelle
		productUpdateQuery = "UPDATE products SET stock = stock - $1 WHERE id = $2"
		_, err = tx.Exec(context.Background(), productUpdateQuery, quantity, productID)

	case "increase":
		// Ürünün ana warehouse'ını bul
		var warehouseID int
		err = tx.QueryRow(context.Background(),
			`SELECT warehouse_id FROM products WHERE id = $1`,
			productID).Scan(&warehouseID)
		if err != nil {
			// Eğer products tablosunda warehouse_id yoksa, ilk bulduğu warehouse'ı kullan
			err = tx.QueryRow(context.Background(),
				`SELECT warehouse_id FROM warehouse_stocks WHERE product_id = $1 LIMIT 1`,
				productID).Scan(&warehouseID)
			if err != nil {
				return err
			}
		}

		updateQuery = `UPDATE warehouse_stocks SET quantity = quantity + $1, updated_at = CURRENT_TIMESTAMP 
                       WHERE product_id = $2 AND warehouse_id = $3`
		_, err = tx.Exec(context.Background(), updateQuery, quantity, productID, warehouseID)

		// Products tablosunu da güncelle
		productUpdateQuery = "UPDATE products SET stock = stock + $1 WHERE id = $2"
		_, err = tx.Exec(context.Background(), productUpdateQuery, quantity, productID)

	default:
		return &InvalidOperationError{Operation: operation}
	}

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// Custom error types
type InsufficientWarehouseStockError struct {
	ProductID      int
	WarehouseID    int
	RequiredStock  int
	AvailableStock int
}

func (e *InsufficientWarehouseStockError) Error() string {
	return "insufficient warehouse stock"
}

type InvalidOperationError struct {
	Operation string
}

func (e *InvalidOperationError) Error() string {
	return "invalid operation"
}
