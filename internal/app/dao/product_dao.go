package dao

import (
	"context"
	"database/sql"

	"food-ordering-api/internal/domain"
)

// ProductDAO handles database operations for products
type ProductDAO struct {
	db *sql.DB
}

// NewProductDAO creates a new ProductDAO
func NewProductDAO(db *sql.DB) domain.ProductDAO {
	return &ProductDAO{db: db}
}

// GetAll retrieves all products from the database
func (dao *ProductDAO) GetAll(ctx context.Context) ([]domain.Product, error) {
	query := `SELECT id, name, price, category FROM products`
	rows, err := dao.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Category); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// GetByID retrieves a product by its ID
func (dao *ProductDAO) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `SELECT id, name, price, category FROM products WHERE id = ?`
	var p domain.Product
	err := dao.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Category)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
