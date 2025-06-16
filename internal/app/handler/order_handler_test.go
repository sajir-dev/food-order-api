package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"food-ordering-api/internal/domain"
	"food-ordering-api/internal/domain/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestFiles(t *testing.T) (string, func()) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coupon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test files with known content
	testFiles := map[string]string{
		"couponbase1": "TEST12345\nVALID1234\nINVALID12",
		"couponbase2": "TEST12345\nVALID1234\nOTHER123",
		"couponbase3": "TEST12345\nOTHER123\nANOTHER1",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestValidateCouponInFiles(t *testing.T) {
	// Setup test files
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	handler := &OrderHandler{}

	// Temporarily change the coupons directory for testing
	originalDir := "coupons"
	os.Rename(originalDir, originalDir+"_backup")
	os.Rename(tempDir, originalDir)
	defer func() {
		os.Rename(originalDir, tempDir)
		os.Rename(originalDir+"_backup", originalDir)
	}()

	tests := []struct {
		name        string
		couponCode  string
		wantValid   bool
		wantErr     bool
		description string
	}{
		{
			name:        "Valid coupon in all files",
			couponCode:  "TEST12345",
			wantValid:   true,
			wantErr:     false,
			description: "Coupon exists in all three files",
		},
		{
			name:        "Valid coupon in two files",
			couponCode:  "VALID1234",
			wantValid:   true,
			wantErr:     false,
			description: "Coupon exists in exactly two files",
		},
		{
			name:        "Invalid coupon in one file",
			couponCode:  "INVALID12",
			wantValid:   false,
			wantErr:     false,
			description: "Coupon exists in only one file",
		},
		{
			name:        "Non-existent coupon",
			couponCode:  "NONEXIST",
			wantValid:   false,
			wantErr:     false,
			description: "Coupon doesn't exist in any file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := handler.validateCouponInFiles(tt.couponCode)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.description)
			}

			assert.Equal(t, tt.wantValid, valid, "Validation result mismatch for test case: %s", tt.description)
		})
	}
}

func TestValidateCouponInFiles_ErrorCases(t *testing.T) {
	// Create handler with test logger
	handler := &OrderHandler{}

	// Test with non-existent directory
	os.Rename("coupons", "coupons_backup")
	defer os.Rename("coupons_backup", "coupons")

	valid, err := handler.validateCouponInFiles("TEST12345")
	assert.Error(t, err, "Expected error when directory doesn't exist")
	assert.False(t, valid, "Should return false when error occurs")
}

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

	// Create test coupon files
	couponDir := "coupons"
	err := os.MkdirAll(couponDir, 0755)
	assert.NoError(t, err)

	// Create test coupon files with valid coupon
	validCoupon := "HAPPYHOURS"
	for _, filename := range []string{"couponbase1", "couponbase2", "couponbase3"} {
		filePath := filepath.Join(couponDir, filename)
		err := os.WriteFile(filePath, []byte(validCoupon), 0644)
		assert.NoError(t, err)
	}

	// Clean up test files after test
	defer os.RemoveAll(couponDir)

	tests := []struct {
		name           string
		requestBody    OrderRequest
		mockOrder      *domain.Order
		mockError      error
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
				CouponCode: validCoupon,
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

			// Create handler with mock
			handler := NewOrderHandler(mockLogic)
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
