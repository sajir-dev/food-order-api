package handler

import (
	"net/http"
	"strconv"

	"food-ordering-api/internal/app/logic"
	"food-ordering-api/internal/domain"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	logic      domain.OrderLogic
	promoLogic logic.PromoLogic
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(logic domain.OrderLogic, promoLogic logic.PromoLogic) *OrderHandler {
	return &OrderHandler{
		logic:      logic,
		promoLogic: promoLogic,
	}
}

// OrderRequest represents the request payload for creating orders
type OrderRequest struct {
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}

// OrderItem represents individual items in an order
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// PlaceOrder handles POST /api/order requests
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to decode request body", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Order must contain at least one item",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Validate each item's quantity
	for i, item := range req.Items {
		if item.Quantity < 1 {
			details := "Key: 'OrderRequest.Items[" + strconv.Itoa(i) + "].Quantity' Error:Field validation for 'Quantity' failed on the 'min' tag"
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid request body",
				Code:    http.StatusBadRequest,
				Details: details,
			})
			return
		}
	}

	// Validate coupon code length if provided
	if req.CouponCode != "" {
		couponLength := len(req.CouponCode)
		if couponLength < 8 || couponLength > 10 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Coupon code must be between 8 and 10 characters",
				Code:  http.StatusBadRequest,
			})
			return
		}

		// Validate coupon using PromoLogic
		valid, err := h.promoLogic.ValidatePromo(req.CouponCode)
		if err != nil {
			logger.Error("failed to validate coupon", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to validate coupon",
				Code:    http.StatusInternalServerError,
				Details: err.Error(),
			})
			return
		}

		if !valid {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid coupon code",
				Code:  http.StatusBadRequest,
			})
			return
		}
	}

	// Convert request to domain model
	orderReq := domain.OrderRequest{
		CouponCode: req.CouponCode,
		Items:      make([]domain.OrderItem, len(req.Items)),
	}

	for i, item := range req.Items {
		orderReq.Items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Create order
	order, err := h.logic.PlaceOrder(c.Request.Context(), orderReq)
	if err != nil {
		logger.Error("failed to place order", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to place order",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Data: order,
	})
}
