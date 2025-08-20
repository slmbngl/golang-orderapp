package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/slmbngl/OrderAplication/internal/adapters/db"
	"github.com/slmbngl/OrderAplication/internal/models"
)

type WarehouseRepository interface {
	// Warehouse management
	CreateWarehouse(warehouse *models.CreateWarehouseRequest) (*models.Warehouse, error)
	GetAllWarehouses() ([]models.Warehouse, error)
	GetWarehouseByID(id int) (*models.Warehouse, error)
	UpdateWarehouse(id int, warehouse *models.UpdateWarehouseRequest) error
	DeleteWarehouse(id int) error

	// Stock management
	GetWarehouseStocks(warehouseID int) ([]models.WarehouseStock, error)
	GetProductStockInWarehouse(warehouseID, productID int) (*models.WarehouseStock, error)
	GetAllStocks() ([]models.WarehouseStock, error)
	UpdateStock(warehouseID, productID, quantity int) error
	AddStock(warehouseID, productID, quantity int) error

	// Transfer management
	CreateStockTransfer(transfer *models.StockTransferRequest, requestedBy int) (*models.StockTransfer, error)
	GetAllTransfers() ([]models.StockTransfer, error)
	GetTransferByID(id int) (*models.StockTransfer, error)
	UpdateTransferStatus(id int, status string) error
	ProcessTransfer(id int) error
}

type warehouseRepo struct{}

func NewWarehouseRepository() WarehouseRepository {
	return &warehouseRepo{}
}

// Warehouse management
func (r *warehouseRepo) CreateWarehouse(req *models.CreateWarehouseRequest) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO warehouses (name, address) VALUES ($1, $2) 
         RETURNING id, name, address, is_active, created_at`,
		req.Name, req.Address).Scan(&warehouse.ID, &warehouse.Name, &warehouse.Address,
		&warehouse.IsActive, &warehouse.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &warehouse, nil
}

func (r *warehouseRepo) GetAllWarehouses() ([]models.Warehouse, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, name, address, is_active, created_at FROM warehouses ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		err := rows.Scan(&w.ID, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt)
		if err != nil {
			return nil, err
		}
		warehouses = append(warehouses, w)
	}

	return warehouses, nil
}

func (r *warehouseRepo) GetWarehouseByID(id int) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, name, address, is_active, created_at FROM warehouses WHERE id = $1`,
		id).Scan(&warehouse.ID, &warehouse.Name, &warehouse.Address, &warehouse.IsActive, &warehouse.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &warehouse, nil
}

func (r *warehouseRepo) UpdateWarehouse(id int, req *models.UpdateWarehouseRequest) error {
	result, err := db.Pool.Exec(context.Background(),
		`UPDATE warehouses SET name = $1, address = $2, is_active = $3 WHERE id = $4`,
		req.Name, req.Address, req.IsActive, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *warehouseRepo) DeleteWarehouse(id int) error {
	// Check if warehouse has stock
	var stockCount int
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM warehouse_stocks WHERE warehouse_id = $1 AND quantity > 0`,
		id).Scan(&stockCount)
	if err != nil {
		return err
	}

	if stockCount > 0 {
		return &WarehouseHasStockError{WarehouseID: id}
	}

	result, err := db.Pool.Exec(context.Background(),
		`DELETE FROM warehouses WHERE id = $1`, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Stock management
func (r *warehouseRepo) GetWarehouseStocks(warehouseID int) ([]models.WarehouseStock, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT ws.id, ws.warehouse_id, ws.product_id, ws.quantity, ws.reserved_quantity,
                ws.created_at, ws.updated_at, w.name, p.name, p.price
         FROM warehouse_stocks ws
         JOIN warehouses w ON ws.warehouse_id = w.id
         JOIN products p ON ws.product_id = p.id
         WHERE ws.warehouse_id = $1
         ORDER BY p.name`, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.WarehouseStock
	for rows.Next() {
		var stock models.WarehouseStock
		err := rows.Scan(&stock.ID, &stock.WarehouseID, &stock.ProductID, &stock.Quantity,
			&stock.ReservedQuantity, &stock.CreatedAt, &stock.UpdatedAt,
			&stock.WarehouseName, &stock.ProductName, &stock.ProductPrice)
		if err != nil {
			return nil, err
		}
		stock.AvailableStock = stock.Quantity - stock.ReservedQuantity
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (r *warehouseRepo) GetAllStocks() ([]models.WarehouseStock, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT ws.id, ws.warehouse_id, ws.product_id, ws.quantity, ws.reserved_quantity,
                ws.created_at, ws.updated_at, w.name, p.name, p.price
         FROM warehouse_stocks ws
         JOIN warehouses w ON ws.warehouse_id = w.id
         JOIN products p ON ws.product_id = p.id
         ORDER BY w.name, p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.WarehouseStock
	for rows.Next() {
		var stock models.WarehouseStock
		err := rows.Scan(&stock.ID, &stock.WarehouseID, &stock.ProductID, &stock.Quantity,
			&stock.ReservedQuantity, &stock.CreatedAt, &stock.UpdatedAt,
			&stock.WarehouseName, &stock.ProductName, &stock.ProductPrice)
		if err != nil {
			return nil, err
		}
		stock.AvailableStock = stock.Quantity - stock.ReservedQuantity
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (r *warehouseRepo) GetProductStockInWarehouse(warehouseID, productID int) (*models.WarehouseStock, error) {
	var stock models.WarehouseStock
	err := db.Pool.QueryRow(context.Background(),
		`SELECT ws.id, ws.warehouse_id, ws.product_id, ws.quantity, ws.reserved_quantity,
                ws.created_at, ws.updated_at, w.name, p.name, p.price
         FROM warehouse_stocks ws
         JOIN warehouses w ON ws.warehouse_id = w.id
         JOIN products p ON ws.product_id = p.id
         WHERE ws.warehouse_id = $1 AND ws.product_id = $2`,
		warehouseID, productID).Scan(&stock.ID, &stock.WarehouseID, &stock.ProductID,
		&stock.Quantity, &stock.ReservedQuantity, &stock.CreatedAt, &stock.UpdatedAt,
		&stock.WarehouseName, &stock.ProductName, &stock.ProductPrice)

	if err != nil {
		return nil, err
	}

	stock.AvailableStock = stock.Quantity - stock.ReservedQuantity
	return &stock, nil
}

func (r *warehouseRepo) UpdateStock(warehouseID, productID, quantity int) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Check if stock record exists
	var exists bool
	err = tx.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM warehouse_stocks WHERE warehouse_id = $1 AND product_id = $2)`,
		warehouseID, productID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing stock
		_, err = tx.Exec(context.Background(),
			`UPDATE warehouse_stocks SET quantity = $1, updated_at = CURRENT_TIMESTAMP 
             WHERE warehouse_id = $2 AND product_id = $3`,
			quantity, warehouseID, productID)
	} else {
		// Insert new stock record
		_, err = tx.Exec(context.Background(),
			`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity) 
             VALUES ($1, $2, $3)`,
			warehouseID, productID, quantity)
	}

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (r *warehouseRepo) AddStock(warehouseID, productID, quantity int) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Check if stock record exists
	var exists bool
	err = tx.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM warehouse_stocks WHERE warehouse_id = $1 AND product_id = $2)`,
		warehouseID, productID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Add to existing stock
		_, err = tx.Exec(context.Background(),
			`UPDATE warehouse_stocks SET quantity = quantity + $1, updated_at = CURRENT_TIMESTAMP 
             WHERE warehouse_id = $2 AND product_id = $3`,
			quantity, warehouseID, productID)
	} else {
		// Insert new stock record
		_, err = tx.Exec(context.Background(),
			`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity) 
             VALUES ($1, $2, $3)`,
			warehouseID, productID, quantity)
	}

	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// Transfer management
func (r *warehouseRepo) CreateStockTransfer(req *models.StockTransferRequest, requestedBy int) (*models.StockTransfer, error) {
	var transfer models.StockTransfer
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO stock_transfers (from_warehouse_id, to_warehouse_id, product_id, quantity, reason, requested_by)
         VALUES ($1, $2, $3, $4, $5, $6)
         RETURNING id, from_warehouse_id, to_warehouse_id, product_id, quantity, status, reason, requested_by, created_at, completed_at`,
		req.FromWarehouseID, req.ToWarehouseID, req.ProductID, req.Quantity, req.Reason, requestedBy).
		Scan(&transfer.ID, &transfer.FromWarehouseID, &transfer.ToWarehouseID, &transfer.ProductID,
			&transfer.Quantity, &transfer.Status, &transfer.Reason, &transfer.RequestedBy,
			&transfer.CreatedAt, &transfer.CompletedAt)

	if err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (r *warehouseRepo) GetAllTransfers() ([]models.StockTransfer, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT st.id, st.from_warehouse_id, st.to_warehouse_id, st.product_id, st.quantity,
                st.status, st.reason, st.requested_by, st.created_at, st.completed_at,
                COALESCE(wf.name, 'External') as from_warehouse_name,
                COALESCE(wt.name, 'External') as to_warehouse_name,
                p.name as product_name, u.username as requested_by_user
         FROM stock_transfers st
         LEFT JOIN warehouses wf ON st.from_warehouse_id = wf.id
         LEFT JOIN warehouses wt ON st.to_warehouse_id = wt.id
         JOIN products p ON st.product_id = p.id
         JOIN users u ON st.requested_by = u.id
         ORDER BY st.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []models.StockTransfer
	for rows.Next() {
		var transfer models.StockTransfer
		err := rows.Scan(&transfer.ID, &transfer.FromWarehouseID, &transfer.ToWarehouseID,
			&transfer.ProductID, &transfer.Quantity, &transfer.Status, &transfer.Reason,
			&transfer.RequestedBy, &transfer.CreatedAt, &transfer.CompletedAt,
			&transfer.FromWarehouseName, &transfer.ToWarehouseName,
			&transfer.ProductName, &transfer.RequestedByUser)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

func (r *warehouseRepo) GetTransferByID(id int) (*models.StockTransfer, error) {
	var transfer models.StockTransfer
	err := db.Pool.QueryRow(context.Background(),
		`SELECT st.id, st.from_warehouse_id, st.to_warehouse_id, st.product_id, st.quantity,
                st.status, st.reason, st.requested_by, st.created_at, st.completed_at,
                COALESCE(wf.name, 'External') as from_warehouse_name,
                COALESCE(wt.name, 'External') as to_warehouse_name,
                p.name as product_name, u.username as requested_by_user
         FROM stock_transfers st
         LEFT JOIN warehouses wf ON st.from_warehouse_id = wf.id
         LEFT JOIN warehouses wt ON st.to_warehouse_id = wt.id
         JOIN products p ON st.product_id = p.id
         JOIN users u ON st.requested_by = u.id
         WHERE st.id = $1`, id).
		Scan(&transfer.ID, &transfer.FromWarehouseID, &transfer.ToWarehouseID,
			&transfer.ProductID, &transfer.Quantity, &transfer.Status, &transfer.Reason,
			&transfer.RequestedBy, &transfer.CreatedAt, &transfer.CompletedAt,
			&transfer.FromWarehouseName, &transfer.ToWarehouseName,
			&transfer.ProductName, &transfer.RequestedByUser)

	if err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (r *warehouseRepo) UpdateTransferStatus(id int, status string) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get current transfer status
	var currentStatus string
	err = tx.QueryRow(context.Background(),
		`SELECT status FROM stock_transfers WHERE id = $1`,
		id).Scan(&currentStatus)
	if err != nil {
		return err
	}

	// If status is the same, no need to update
	if currentStatus == status {
		return tx.Commit(context.Background())
	}

	var completedAt *time.Time
	if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		completedAt = &now
	}

	// Update transfer status
	result, err := tx.Exec(context.Background(),
		`UPDATE stock_transfers SET status = $1, completed_at = $2 WHERE id = $3`,
		status, completedAt, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	// If status is being set to "completed", process the transfer automatically
	if status == "completed" && currentStatus == "pending" {
		// Get transfer details with lock
		var transfer models.StockTransfer
		err = tx.QueryRow(context.Background(),
			`SELECT id, from_warehouse_id, to_warehouse_id, product_id, quantity, status
             FROM stock_transfers WHERE id = $1 FOR UPDATE`,
			id).Scan(&transfer.ID, &transfer.FromWarehouseID, &transfer.ToWarehouseID,
			&transfer.ProductID, &transfer.Quantity, &transfer.Status)
		if err != nil {
			return err
		}

		// Handle stock decrease from source warehouse
		if transfer.FromWarehouseID != nil {
			var currentStock int
			err = tx.QueryRow(context.Background(),
				`SELECT quantity FROM warehouse_stocks 
                 WHERE warehouse_id = $1 AND product_id = $2 FOR UPDATE`,
				*transfer.FromWarehouseID, transfer.ProductID).Scan(&currentStock)
			if err != nil {
				if err == pgx.ErrNoRows {
					return &InsufficientStockError{
						WarehouseID: *transfer.FromWarehouseID,
						ProductID:   transfer.ProductID,
						Required:    transfer.Quantity,
						Available:   0,
					}
				}
				return err
			}

			if currentStock < transfer.Quantity {
				return &InsufficientStockError{
					WarehouseID: *transfer.FromWarehouseID,
					ProductID:   transfer.ProductID,
					Required:    transfer.Quantity,
					Available:   currentStock,
				}
			}

			// Decrease stock from source warehouse
			_, err = tx.Exec(context.Background(),
				`UPDATE warehouse_stocks SET quantity = quantity - $1, updated_at = CURRENT_TIMESTAMP
                 WHERE warehouse_id = $2 AND product_id = $3`,
				transfer.Quantity, *transfer.FromWarehouseID, transfer.ProductID)
			if err != nil {
				return err
			}

			// Also update products table
			_, err = tx.Exec(context.Background(),
				`UPDATE products SET stock = stock - $1 WHERE id = $2`,
				transfer.Quantity, transfer.ProductID)
			if err != nil {
				return err
			}
		}

		// Handle stock increase to destination warehouse
		if transfer.ToWarehouseID != nil {
			// Check if stock record exists for destination
			var exists bool
			err = tx.QueryRow(context.Background(),
				`SELECT EXISTS(SELECT 1 FROM warehouse_stocks WHERE warehouse_id = $1 AND product_id = $2)`,
				*transfer.ToWarehouseID, transfer.ProductID).Scan(&exists)
			if err != nil {
				return err
			}

			if exists {
				// Add to existing stock
				_, err = tx.Exec(context.Background(),
					`UPDATE warehouse_stocks SET quantity = quantity + $1, updated_at = CURRENT_TIMESTAMP
                     WHERE warehouse_id = $2 AND product_id = $3`,
					transfer.Quantity, *transfer.ToWarehouseID, transfer.ProductID)
			} else {
				// Create new stock record
				_, err = tx.Exec(context.Background(),
					`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity)
                     VALUES ($1, $2, $3)`,
					*transfer.ToWarehouseID, transfer.ProductID, transfer.Quantity)
			}
			if err != nil {
				return err
			}

			// Also update products table
			_, err = tx.Exec(context.Background(),
				`UPDATE products SET stock = stock + $1 WHERE id = $2`,
				transfer.Quantity, transfer.ProductID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(context.Background())
}

func (r *warehouseRepo) ProcessTransfer(id int) error {
	// Begin transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get transfer details with lock
	var transfer models.StockTransfer
	err = tx.QueryRow(context.Background(),
		`SELECT id, from_warehouse_id, to_warehouse_id, product_id, quantity, status
         FROM stock_transfers WHERE id = $1 FOR UPDATE`,
		id).Scan(&transfer.ID, &transfer.FromWarehouseID, &transfer.ToWarehouseID,
		&transfer.ProductID, &transfer.Quantity, &transfer.Status)
	if err != nil {
		return err
	}

	// Check if transfer is in pending status
	if transfer.Status != "pending" {
		return &TransferNotPendingError{TransferID: id, Status: transfer.Status}
	}

	// Handle stock decrease from source warehouse
	if transfer.FromWarehouseID != nil {
		var currentStock int
		err = tx.QueryRow(context.Background(),
			`SELECT quantity FROM warehouse_stocks 
             WHERE warehouse_id = $1 AND product_id = $2 FOR UPDATE`,
			*transfer.FromWarehouseID, transfer.ProductID).Scan(&currentStock)
		if err != nil {
			if err == pgx.ErrNoRows {
				return &InsufficientStockError{
					WarehouseID: *transfer.FromWarehouseID,
					ProductID:   transfer.ProductID,
					Required:    transfer.Quantity,
					Available:   0,
				}
			}
			return err
		}

		if currentStock < transfer.Quantity {
			return &InsufficientStockError{
				WarehouseID: *transfer.FromWarehouseID,
				ProductID:   transfer.ProductID,
				Required:    transfer.Quantity,
				Available:   currentStock,
			}
		}

		// Decrease stock from source warehouse
		_, err = tx.Exec(context.Background(),
			`UPDATE warehouse_stocks SET quantity = quantity - $1, updated_at = CURRENT_TIMESTAMP
             WHERE warehouse_id = $2 AND product_id = $3`,
			transfer.Quantity, *transfer.FromWarehouseID, transfer.ProductID)
		if err != nil {
			return err
		}

		// Also update products table
		_, err = tx.Exec(context.Background(),
			`UPDATE products SET stock = stock - $1 WHERE id = $2`,
			transfer.Quantity, transfer.ProductID)
		if err != nil {
			return err
		}
	}

	// Handle stock increase to destination warehouse
	if transfer.ToWarehouseID != nil {
		// Check if stock record exists for destination
		var exists bool
		err = tx.QueryRow(context.Background(),
			`SELECT EXISTS(SELECT 1 FROM warehouse_stocks WHERE warehouse_id = $1 AND product_id = $2)`,
			*transfer.ToWarehouseID, transfer.ProductID).Scan(&exists)
		if err != nil {
			return err
		}

		if exists {
			// Add to existing stock
			_, err = tx.Exec(context.Background(),
				`UPDATE warehouse_stocks SET quantity = quantity + $1, updated_at = CURRENT_TIMESTAMP
                 WHERE warehouse_id = $2 AND product_id = $3`,
				transfer.Quantity, *transfer.ToWarehouseID, transfer.ProductID)
		} else {
			// Create new stock record
			_, err = tx.Exec(context.Background(),
				`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity)
                 VALUES ($1, $2, $3)`,
				*transfer.ToWarehouseID, transfer.ProductID, transfer.Quantity)
		}
		if err != nil {
			return err
		}

		// Also update products table
		_, err = tx.Exec(context.Background(),
			`UPDATE products SET stock = stock + $1 WHERE id = $2`,
			transfer.Quantity, transfer.ProductID)
		if err != nil {
			return err
		}
	}

	// Update transfer status to completed
	now := time.Now()
	_, err = tx.Exec(context.Background(),
		`UPDATE stock_transfers SET status = 'completed', completed_at = $1 WHERE id = $2`,
		now, id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// Custom error types
type WarehouseHasStockError struct {
	WarehouseID int
}

func (e *WarehouseHasStockError) Error() string {
	return "warehouse has stock and cannot be deleted"
}

type InsufficientStockError struct {
	WarehouseID int
	ProductID   int
	Required    int
	Available   int
}

func (e *InsufficientStockError) Error() string {
	return "insufficient stock for transfer"
}

type TransferNotPendingError struct {
	TransferID int
	Status     string
}

func (e *TransferNotPendingError) Error() string {
	return "transfer is not in pending status"
}
