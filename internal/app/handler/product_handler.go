package handler

import (
	"log/slog"
	"net/http"

	"food-ordering-api/internal/domain"

	"github.com/gin-gonic/gin"
)

var logger *slog.Logger

// SetLogger sets the package-level logger
func SetLogger(l *slog.Logger) {
	logger = l
}

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	logic domain.ProductLogic
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(logic domain.ProductLogic) *ProductHandler {
	return &ProductHandler{
		logic: logic,
	}
}

// ListProducts handles GET /api/products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.logic.ListProducts(c.Request.Context())
	if err != nil {
		logger.Error("failed to list products", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list products",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Data: products})
}

// GetProduct handles GET /api/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.logic.GetProduct(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get product", "error", err, "productID", id)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get product",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Product not found",
			Code:  http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Data: product})
}
