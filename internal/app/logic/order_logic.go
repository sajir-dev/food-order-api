package logic

import (
	"context"
	"fmt"

	"food-ordering-api/internal/domain"
)

// OrderLogic defines the interface for order business logic
type OrderLogic interface {
	PlaceOrder(ctx context.Context, req OrderRequest) (*Order, error)
}

// Order represents a customer order
type Order struct {
	ID       string      `json:"id" db:"id"`
	Items    []OrderItem `json:"items"`
	Products []Product   `json:"products"`
}

// OrderItem represents individual items in an order
type OrderItem struct {
	ProductID string `json:"productId" db:"product_id"`
	Quantity  int    `json:"quantity" db:"quantity"`
}

// OrderRequest represents the request payload for creating orders
type OrderRequest struct {
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}

// OrderLogicImpl implements the OrderLogic interface
type OrderLogicImpl struct {
	orderDAO   domain.OrderDAO
	productDAO domain.ProductDAO
}

// NewOrderLogic creates a new instance of OrderLogicImpl
func NewOrderLogic(orderDAO domain.OrderDAO, productDAO domain.ProductDAO) domain.OrderLogic {
	return &OrderLogicImpl{
		orderDAO:   orderDAO,
		productDAO: productDAO,
	}
}

// PlaceOrder creates a new order
func (l *OrderLogicImpl) PlaceOrder(ctx context.Context, req domain.OrderRequest) (*domain.Order, error) {
	// Validate request
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must contain at least one item")
	}

	// Validate all products exist and get their details
	var products []domain.Product
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be greater than 0 for product %s", item.ProductID)
		}

		product, err := l.productDAO.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate product %s: %w", item.ProductID, err)
		}
		if product == nil {
			return nil, fmt.Errorf("product not found: %s", item.ProductID)
		}
		products = append(products, *product)
	}

	// Create order
	order := &domain.Order{
		Items:    req.Items,
		Products: products,
	}

	// Save order to database
	if err := l.orderDAO.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}
