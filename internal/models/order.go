package models

import "time"

type Order struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Status      string    `json:"status,omitempty"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	Username    string    `json:"username,omitempty"` // View'dan gelecek
}

type OrderItem struct {
	ID                 int     `json:"id" db:"id"`
	OrderID            int     `json:"order_id" db:"order_id"`
	ProductID          int     `json:"product_id" db:"product_id"`
	Quantity           int     `json:"quantity" db:"quantity"`
	Price              float64 `json:"price" db:"price"`
	ProductName        string  `json:"product_name,omitempty"`
	ProductDescription string  `json:"product_description,omitempty"`
}

// Request structs
type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type OrderWithItems struct {
	Order Order       `json:"order"`
	Items []OrderItem `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" example:"confirmed"`
}

// View related structs
type OrderWithDetails struct {
	Order    Order         `json:"order"`
	Products []ProductInfo `json:"products"`
}

type ProductInfo struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}
