package models

import "time"

type Product struct {
	ID          int       `json:"id" example:"1"`
	Name        string    `json:"name" validate:"required" example:"Laptop"`
	Description string    `json:"description" example:"High performance laptop"`
	Price       float64   `json:"price" validate:"required" example:"999.99"`
	Stock       int       `json:"stock" validate:"required" example:"10"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductRequest struct {
	Name        string  `json:"name" validate:"required" example:"Laptop"`
	Description string  `json:"description" example:"High performance laptop"`
	Price       float64 `json:"price" validate:"required" example:"999.99"`
	Stock       int     `json:"stock" validate:"required" example:"10"`
}
