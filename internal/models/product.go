package models

import "time"

type Product struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required" example:"Laptop"`
	Description string    `json:"description" db:"description" example:"High performance laptop"`
	Price       float64   `json:"price" db:"price" validate:"required" example:"999.99"`
	Stock       int       `json:"stock" db:"stock" validate:"required" example:"10"`
	WarehouseID int       `json:"warehouse_id" db:"warehouse_id" validate:"required"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	WarehouseName string `json:"warehouse_name,omitempty"`
}

type ProductRequest struct {
	Name        string  `json:"name" validate:"required" example:"Laptop"`
	Description string  `json:"description" example:"High performance laptop"`
	Price       float64 `json:"price" validate:"required" example:"999.99"`
	Stock       int     `json:"stock" validate:"required" example:"10"`
	WarehouseID int     `json:"warehouse_id" validate:"required"`
}
