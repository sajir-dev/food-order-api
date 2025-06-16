package logic

import (
	"context"
	"testing"

	"food-ordering-api/internal/domain"
	"food-ordering-api/internal/domain/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductLogicImpl_ListProducts(t *testing.T) {
	tests := []struct {
		name          string
		mockProducts  []domain.Product
		mockError     error
		expectedError bool
		expectedCount int
	}{
		{
			name: "successful product list retrieval",
			mockProducts: []domain.Product{
				{
					ID:       "1",
					Name:     "Pizza",
					Price:    12.99,
					Category: "Italian",
				},
				{
					ID:       "2",
					Name:     "Burger",
					Price:    8.99,
					Category: "Fast Food",
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:          "empty product list",
			mockProducts:  []domain.Product{},
			mockError:     nil,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "error retrieving products",
			mockProducts:  nil,
			mockError:     assert.AnError,
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock ProductDAO
			mockDAO := mocks.NewProductDAO(t)

			// Set up expectations
			mockDAO.On("GetAll", mock.Anything).Return(tt.mockProducts, tt.mockError)

			// Create ProductLogic instance with mock
			logic := NewProductLogic(mockDAO)

			// Call the method being tested
			products, err := logic.ListProducts(context.Background())

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, products)
			} else {
				assert.NoError(t, err)
				assert.Len(t, products, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, tt.mockProducts, products)
				}
			}

			// Verify that all expectations were met
			mockDAO.AssertExpectations(t)
		})
	}
}

func TestProductLogicImpl_GetProduct(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		mockProduct   *domain.Product
		mockError     error
		expectedError bool
	}{
		{
			name:      "successful product retrieval",
			productID: "1",
			mockProduct: &domain.Product{
				ID:       "1",
				Name:     "Pizza",
				Price:    12.99,
				Category: "Italian",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "product not found",
			productID:     "999",
			mockProduct:   nil,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "empty product ID",
			productID:     "",
			mockProduct:   nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "error retrieving product",
			productID:     "1",
			mockProduct:   nil,
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock ProductDAO
			mockDAO := mocks.NewProductDAO(t)

			// Set up expectations only if productID is not empty
			if tt.productID != "" {
				mockDAO.On("GetByID", mock.Anything, tt.productID).Return(tt.mockProduct, tt.mockError)
			}

			// Create ProductLogic instance with mock
			logic := NewProductLogic(mockDAO)

			// Call the method being tested
			product, err := logic.GetProduct(context.Background(), tt.productID)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				if tt.mockProduct != nil {
					assert.Equal(t, tt.mockProduct, product)
				} else {
					assert.Nil(t, product)
				}
			}

			// Verify that all expectations were met
			mockDAO.AssertExpectations(t)
		})
	}
}
