package handler

import (
	"bytes"
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

func setupOrderTestRouter(handler *OrderHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/orders", handler.PlaceOrder)
	return router
}

func TestOrderHandler_PlaceOrder(t *testing.T) {
	// Initialize logger for tests
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	SetLogger(logger)

	tests := []struct {
		name           string
		requestBody    OrderRequest
		mockOrder      *domain.Order
		mockError      error
		mockPromoValid bool
		mockPromoErr   error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "success - order placed without coupon",
			requestBody: OrderRequest{
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockOrder: &domain.Order{
				ID: "order123",
				Items: []domain.OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id": "order123",
					"items": []interface{}{
						map[string]interface{}{
							"productId": "1",
							"quantity":  float64(2),
						},
					},
					"products": nil,
				},
			},
		},
		{
			name: "success - order placed with valid coupon",
			requestBody: OrderRequest{
				CouponCode: "VALID1234",
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockOrder: &domain.Order{
				ID: "order123",
				Items: []domain.OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockError:      nil,
			mockPromoValid: true,
			mockPromoErr:   nil,
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id": "order123",
					"items": []interface{}{
						map[string]interface{}{
							"productId": "1",
							"quantity":  float64(2),
						},
					},
					"products": nil,
				},
			},
		},
		{
			name: "error - empty items",
			requestBody: OrderRequest{
				Items: []OrderItem{},
			},
			mockOrder:      nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Order must contain at least one item",
				"code":  float64(http.StatusBadRequest),
			},
		},
		{
			name: "error - invalid coupon length",
			requestBody: OrderRequest{
				CouponCode: "SHORT",
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockOrder:      nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Coupon code must be between 8 and 10 characters",
				"code":  float64(http.StatusBadRequest),
			},
		},
		{
			name: "error - invalid coupon code",
			requestBody: OrderRequest{
				CouponCode: "INVALID12",
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockOrder:      nil,
			mockError:      nil,
			mockPromoValid: false,
			mockPromoErr:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid coupon code",
				"code":  float64(http.StatusBadRequest),
			},
		},
		{
			name: "error - coupon validation error",
			requestBody: OrderRequest{
				CouponCode: "VALID1234",
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockOrder:      nil,
			mockError:      nil,
			mockPromoValid: false,
			mockPromoErr:   errors.New("failed to read coupon file"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Failed to validate coupon",
				"code":    float64(http.StatusInternalServerError),
				"details": "failed to read coupon file",
			},
		},
		{
			name: "error - invalid request body",
			requestBody: OrderRequest{
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  -1, // Invalid quantity
					},
				},
			},
			mockOrder:      nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid request body",
				"code":    float64(http.StatusBadRequest),
				"details": "Key: 'OrderRequest.Items[0].Quantity' Error:Field validation for 'Quantity' failed on the 'min' tag",
			},
		},
		{
			name: "error - internal server error",
			requestBody: OrderRequest{
				Items: []OrderItem{
					{
						ProductID: "1",
						Quantity:  2,
					},
				},
			},
			mockOrder:      nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":   "Failed to place order",
				"code":    float64(http.StatusInternalServerError),
				"details": "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockLogic := mocks.NewOrderLogic(t)
			if (tt.mockOrder != nil || tt.mockError != nil) && tt.name != "error - invalid request body" {
				mockLogic.On("PlaceOrder", mock.Anything, mock.MatchedBy(func(req domain.OrderRequest) bool {
					return len(req.Items) == len(tt.requestBody.Items) &&
						req.CouponCode == tt.requestBody.CouponCode
				})).Return(tt.mockOrder, tt.mockError)
			}

			// Setup mock PromoLogic
			mockPromoLogic := &mockPromoLogic{
				validatePromoFunc: func(couponCode string) (bool, error) {
					return tt.mockPromoValid, tt.mockPromoErr
				},
			}

			// Create handler with mocks
			handler := NewOrderHandler(mockLogic, mockPromoLogic)
			router := setupOrderTestRouter(handler)

			// Create request body
			reqBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// Verify mock expectations
			mockLogic.AssertExpectations(t)
		})
	}
}

// mockPromoLogic is a mock implementation of the PromoLogic interface
type mockPromoLogic struct {
	validatePromoFunc func(couponCode string) (bool, error)
}

func (m *mockPromoLogic) ValidatePromo(couponCode string) (bool, error) {
	return m.validatePromoFunc(couponCode)
}
