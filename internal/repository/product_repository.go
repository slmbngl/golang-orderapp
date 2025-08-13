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
	CreateProduct(product *models.ProductRequest) (*models.Product, error)
	UpdateProduct(id int, product *models.ProductRequest) error
	DeleteProduct(id int) error
}

type productRepo struct{}

func NewProductRepository() ProductRepository {
	return &productRepo{}
}

func (r *productRepo) GetAllProducts() ([]models.Product, error) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT id, name, description, price, stock, created_at FROM products ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *productRepo) GetProductByID(id int) (*models.Product, error) {
	var product models.Product
	err := db.Pool.QueryRow(context.Background(),
		"SELECT id, name, description, price, stock, created_at FROM products WHERE id = $1", id).
		Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepo) CreateProduct(productReq *models.ProductRequest) (*models.Product, error) {
	var product models.Product
	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
		productReq.Name, productReq.Description, productReq.Price, productReq.Stock).
		Scan(&product.ID, &product.CreatedAt)

	if err != nil {
		return nil, err
	}

	product.Name = productReq.Name
	product.Description = productReq.Description
	product.Price = productReq.Price
	product.Stock = productReq.Stock

	return &product, nil
}

func (r *productRepo) UpdateProduct(id int, productReq *models.ProductRequest) error {
	result, err := db.Pool.Exec(context.Background(),
		"UPDATE products SET name=$1, description=$2, price=$3, stock=$4 WHERE id=$5",
		productReq.Name, productReq.Description, productReq.Price, productReq.Stock, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
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
