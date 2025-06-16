package logic

import (
	"context"
	"testing"

	"food-ordering-api/internal/domain"
	"food-ordering-api/internal/domain/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderLogicImpl_PlaceOrder(t *testing.T) {
	// Sample test data
	validProduct := &domain.Product{
		ID:       "1",
		Name:     "Pizza",
		Price:    12.99,
		Category: "Italian",
	}

	validOrderRequest := domain.OrderRequest{
		Items: []domain.OrderItem{
			{
				ProductID: "1",
				Quantity:  2,
			},
		},
	}

	tests := []struct {
		name           string
		request        domain.OrderRequest
		mockProduct    *domain.Product
		mockProductErr error
		mockOrderErr   error
		expectedError  bool
		errorMessage   string
	}{
		{
			name:           "successful order placement",
			request:        validOrderRequest,
			mockProduct:    validProduct,
			mockProductErr: nil,
			mockOrderErr:   nil,
			expectedError:  false,
		},
		{
			name: "empty order items",
			request: domain.OrderRequest{
				Items: []domain.OrderItem{},
			},
			mockProduct:    nil,
			mockProductErr: nil,
			mockOrderErr:   nil,
			expectedError:  true,
			errorMessage:   "order must contain at least one item",
		},
		{
			name: "invalid quantity",
			request: domain.OrderRequest{
				Items: []domain.OrderItem{
					{
						ProductID: "1",
						Quantity:  0,
					},
				},
			},
			mockProduct:    nil,
			mockProductErr: nil,
			mockOrderErr:   nil,
			expectedError:  true,
			errorMessage:   "quantity must be greater than 0",
		},
		{
			name:           "product not found",
			request:        validOrderRequest,
			mockProduct:    nil,
			mockProductErr: nil,
			mockOrderErr:   nil,
			expectedError:  true,
			errorMessage:   "product not found",
		},
		{
			name:           "product validation error",
			request:        validOrderRequest,
			mockProduct:    nil,
			mockProductErr: assert.AnError,
			mockOrderErr:   nil,
			expectedError:  true,
			errorMessage:   "failed to validate product",
		},
		{
			name:           "order creation error",
			request:        validOrderRequest,
			mockProduct:    validProduct,
			mockProductErr: nil,
			mockOrderErr:   assert.AnError,
			expectedError:  true,
			errorMessage:   "failed to create order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockOrderDAO := mocks.NewOrderDAO(t)
			mockProductDAO := mocks.NewProductDAO(t)

			// Set up expectations
			if len(tt.request.Items) > 0 {
				// Only set up product validation expectations if we expect to reach that point
				// Skip for invalid quantity case since it should fail before product validation
				if tt.name != "invalid quantity" {
					for _, item := range tt.request.Items {
						mockProductDAO.On("GetByID", mock.Anything, item.ProductID).
							Return(tt.mockProduct, tt.mockProductErr)
					}
				}
			}

			if !tt.expectedError || (tt.expectedError && tt.mockProductErr == nil && tt.mockProduct != nil) {
				mockOrderDAO.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).
					Return(tt.mockOrderErr)
			}

			// Create OrderLogic instance with mocks
			logic := NewOrderLogic(mockOrderDAO, mockProductDAO)

			// Call the method being tested
			order, err := logic.PlaceOrder(context.Background(), tt.request)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, order)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.request.Items, order.Items)
				assert.Equal(t, []domain.Product{*validProduct}, order.Products)
			}

			// Verify that all expectations were met
			mockOrderDAO.AssertExpectations(t)
			mockProductDAO.AssertExpectations(t)
		})
	}
}
