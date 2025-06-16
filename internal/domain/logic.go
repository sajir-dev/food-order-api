package domain

import "context"

// ProductLogic defines the interface for product business logic operations
type ProductLogic interface {
	ListProducts(ctx context.Context) ([]Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
}

// OrderLogic defines the interface for order business logic operations
type OrderLogic interface {
	PlaceOrder(ctx context.Context, req OrderRequest) (*Order, error)
}

// OrderRequest represents the request payload for creating orders
type OrderRequest struct {
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}
