package domain

import "context"

// ProductDAO defines the interface for product data access operations
type ProductDAO interface {
	GetAll(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id string) (*Product, error)
}

// OrderDAO defines the interface for order data access operations
type OrderDAO interface {
	Create(ctx context.Context, order *Order) error
	GetOrderItems(ctx context.Context, orderID string) ([]OrderItem, error)
}
