package logic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestPromoLogic_ValidatePromo(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	promoLogic := NewPromoLogic(tempDir, []string{"couponbase1", "couponbase2", "couponbase3"})

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
		{
			name:        "Empty coupon code",
			couponCode:  "",
			wantValid:   false,
			wantErr:     false,
			description: "Empty coupon code should be invalid",
		},
		{
			name:        "Coupon with whitespace",
			couponCode:  " TEST12345 ",
			wantValid:   false,
			wantErr:     false,
			description: "Coupon with whitespace should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := promoLogic.ValidatePromo(tt.couponCode)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.description)
			}

			assert.Equal(t, tt.wantValid, valid, "Validation result mismatch for test case: %s", tt.description)
		})
	}
}

func TestPromoLogic_ValidatePromo_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, func())
		couponCode  string
		wantValid   bool
		wantErr     bool
		description string
	}{
		{
			name: "Non-existent directory",
			setup: func(t *testing.T) (string, func()) {
				tempDir, err := os.MkdirTemp("", "coupon_test_*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				cleanup := func() {
					os.RemoveAll(tempDir)
				}
				return tempDir, cleanup
			},
			couponCode:  "TEST12345",
			wantValid:   false,
			wantErr:     true,
			description: "Should error when directory doesn't exist",
		},
		{
			name: "Empty file list",
			setup: func(t *testing.T) (string, func()) {
				tempDir, err := os.MkdirTemp("", "coupon_test_*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				cleanup := func() {
					os.RemoveAll(tempDir)
				}
				return tempDir, cleanup
			},
			couponCode:  "TEST12345",
			wantValid:   false,
			wantErr:     false,
			description: "Should return false when no files to check",
		},
		{
			name: "Unreadable file",
			setup: func(t *testing.T) (string, func()) {
				tempDir, err := os.MkdirTemp("", "coupon_test_*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}

				// Create a file with no read permissions
				filePath := filepath.Join(tempDir, "couponbase1")
				if err := os.WriteFile(filePath, []byte("TEST12345"), 0000); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				cleanup := func() {
					os.RemoveAll(tempDir)
				}
				return tempDir, cleanup
			},
			couponCode:  "TEST12345",
			wantValid:   false,
			wantErr:     true,
			description: "Should error when file is unreadable",
		},
		{
			name: "File too large",
			setup: func(t *testing.T) (string, func()) {
				tempDir, err := os.MkdirTemp("", "coupon_test_*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}

				// Create a file with a very long line
				filePath := filepath.Join(tempDir, "couponbase1")
				longLine := make([]byte, 11*1024*1024) // 11MB line
				if err := os.WriteFile(filePath, longLine, 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				cleanup := func() {
					os.RemoveAll(tempDir)
				}
				return tempDir, cleanup
			},
			couponCode:  "TEST12345",
			wantValid:   false,
			wantErr:     true,
			description: "Should error when file line is too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := tt.setup(t)
			defer cleanup()

			var files []string
			if tt.name == "Empty file list" {
				files = []string{} // Empty file list for this test case
			} else {
				files = []string{"couponbase1"}
			}
			promoLogic := NewPromoLogic(tempDir, files)
			valid, err := promoLogic.ValidatePromo(tt.couponCode)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.description)
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.description)
			}

			assert.Equal(t, tt.wantValid, valid, "Validation result mismatch for test case: %s", tt.description)
		})
	}
}
