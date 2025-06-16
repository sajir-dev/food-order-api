package logic

import (
	"context"
	"fmt"

	"food-ordering-api/internal/domain"
)

// ProductLogic defines the interface for product business logic
type ProductLogic interface {
	ListProducts(ctx context.Context) ([]Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
}

// Product represents a food item
type Product struct {
	ID       string  `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	Price    float64 `json:"price" db:"price"`
	Category string  `json:"category" db:"category"`
}

// ProductLogicImpl implements the ProductLogic interface
type ProductLogicImpl struct {
	productDAO domain.ProductDAO
}

// NewProductLogic creates a new instance of ProductLogicImpl
func NewProductLogic(productDAO domain.ProductDAO) domain.ProductLogic {
	return &ProductLogicImpl{
		productDAO: productDAO,
	}
}

// ListProducts retrieves all products
func (l *ProductLogicImpl) ListProducts(ctx context.Context) ([]domain.Product, error) {
	products, err := l.productDAO.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}

// GetProduct retrieves a single product by ID
func (l *ProductLogicImpl) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	product, err := l.productDAO.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return product, nil
}
