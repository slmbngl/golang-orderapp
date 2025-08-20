package models

import "time"

type Warehouse struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required"`
	Address   string    `json:"address" db:"address"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WarehouseStock struct {
	ID               int       `json:"id" db:"id"`
	WarehouseID      int       `json:"warehouse_id" db:"warehouse_id"`
	ProductID        int       `json:"product_id" db:"product_id"`
	Quantity         int       `json:"quantity" db:"quantity"`
	ReservedQuantity int       `json:"reserved_quantity" db:"reserved_quantity"`
	AvailableStock   int       `json:"available_stock"` // Calculated: quantity - reserved_quantity
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	WarehouseName string  `json:"warehouse_name,omitempty"`
	ProductName   string  `json:"product_name,omitempty"`
	ProductPrice  float64 `json:"product_price,omitempty"`
}

type StockTransfer struct {
	ID              int        `json:"id" db:"id"`
	FromWarehouseID *int       `json:"from_warehouse_id" db:"from_warehouse_id"`
	ToWarehouseID   *int       `json:"to_warehouse_id" db:"to_warehouse_id"`
	ProductID       int        `json:"product_id" db:"product_id"`
	Quantity        int        `json:"quantity" db:"quantity"`
	Status          string     `json:"status" db:"status"`
	Reason          string     `json:"reason" db:"reason"`
	RequestedBy     int        `json:"requested_by" db:"requested_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`

	// Joined fields
	FromWarehouseName string `json:"from_warehouse_name,omitempty"`
	ToWarehouseName   string `json:"to_warehouse_name,omitempty"`
	ProductName       string `json:"product_name,omitempty"`
	RequestedByUser   string `json:"requested_by_user,omitempty"`
}

// Request models
type CreateWarehouseRequest struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address"`
}

type UpdateWarehouseRequest struct {
	Name     string `json:"name" validate:"required"`
	Address  string `json:"address"`
	IsActive bool   `json:"is_active"`
}

type StockTransferRequest struct {
	FromWarehouseID *int   `json:"from_warehouse_id"`
	ToWarehouseID   *int   `json:"to_warehouse_id"`
	ProductID       int    `json:"product_id" validate:"required"`
	Quantity        int    `json:"quantity" validate:"required,min=1"`
	Reason          string `json:"reason"`
}

type UpdateStockRequest struct {
	Quantity int    `json:"quantity" validate:"required,min=0"`
	Reason   string `json:"reason"`
}

type StockTransferStatusRequest struct {
	Status string `json:"status" validate:"required"`
}
