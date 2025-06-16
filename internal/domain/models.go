package domain

// Product represents a food item
type Product struct {
	ID       string  `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	Price    float64 `json:"price" db:"price"`
	Category string  `json:"category" db:"category"`
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
