package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"food-ordering-api/internal/domain"
	"food-ordering-api/internal/domain/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter(handler *ProductHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products", handler.ListProducts)
	router.GET("/api/products/:id", handler.GetProduct)
	return router
}

func TestProductHandler_ListProducts(t *testing.T) {
	// Initialize logger for tests
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	SetLogger(logger)

	tests := []struct {
		name           string
		mockProducts   []domain.Product
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "success - returns list of products",
			mockProducts: []domain.Product{
				{ID: "1", Name: "Chicken Waffle", Price: 12.99, Category: "Waffle"},
				{ID: "2", Name: "Belgian Waffle", Price: 9.99, Category: "Waffle"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":       "1",
						"name":     "Chicken Waffle",
						"price":    12.99,
						"category": "Waffle",
					},
					map[string]interface{}{
						"id":       "2",
						"name":     "Belgian Waffle",
						"price":    9.99,
						"category": "Waffle",
					},
				},
			},
		},
		{
			name:           "success - returns empty list",
			mockProducts:   []domain.Product{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"data": []interface{}{},
			},
		},
		{
			name:           "error - internal server error",
			mockProducts:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Failed to list products",
				"code":    float64(http.StatusInternalServerError),
				"details": "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockLogic := mocks.NewProductLogic(t)
			mockLogic.On("ListProducts", mock.Anything).Return(tt.mockProducts, tt.mockError)

			// Create handler with mock
			handler := NewProductHandler(mockLogic)
			router := setupTestRouter(handler)

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/products", nil)

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// Verify mock expectations
			mockLogic.AssertExpectations(t)
		})
	}
}

func TestProductHandler_GetProduct(t *testing.T) {
	// Initialize logger for tests
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	SetLogger(logger)

	tests := []struct {
		name           string
		productID      string
		mockProduct    *domain.Product
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "success - returns product",
			productID: "1",
			mockProduct: &domain.Product{
				ID:       "1",
				Name:     "Chicken Waffle",
				Price:    12.99,
				Category: "Waffle",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       "1",
					"name":     "Chicken Waffle",
					"price":    12.99,
					"category": "Waffle",
				},
			},
		},
		{
			name:           "error - product not found",
			productID:      "999",
			mockProduct:    nil,
			mockError:      nil,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Product not found",
				"code":  float64(http.StatusNotFound),
			},
		},
		{
			name:           "error - internal server error",
			productID:      "1",
			mockProduct:    nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Failed to get product",
				"code":    float64(http.StatusInternalServerError),
				"details": "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockLogic := mocks.NewProductLogic(t)
			mockLogic.On("GetProduct", mock.Anything, tt.productID).Return(tt.mockProduct, tt.mockError)

			// Create handler with mock
			handler := NewProductHandler(mockLogic)
			router := setupTestRouter(handler)

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/products/"+tt.productID, nil)

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// Verify mock expectations
			mockLogic.AssertExpectations(t)
		})
	}
}
