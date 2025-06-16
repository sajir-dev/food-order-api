package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"food-ordering-api/internal/domain"
)

// OrderDAO handles database operations for orders
type OrderDAO struct {
	db *sql.DB
}

// NewOrderDAO creates a new OrderDAO
func NewOrderDAO(db *sql.DB) domain.OrderDAO {
	return &OrderDAO{db: db}
}

// Create creates a new order in the database
func (dao *OrderDAO) Create(ctx context.Context, order *domain.Order) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	order.ID = uuid.New().String()
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders (id) VALUES (?)",
		order.ID,
	)
	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)",
			order.ID, item.ProductID, item.Quantity,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetOrderItems retrieves all items for an order
func (dao *OrderDAO) GetOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	query := `SELECT product_id, quantity FROM order_items WHERE order_id = ?`
	rows, err := dao.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
