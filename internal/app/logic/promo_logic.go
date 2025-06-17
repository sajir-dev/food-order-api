package logic

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type promoLogic struct {
	couponDir string
	files     []string
}

func NewPromoLogic(couponDir string, files []string) *promoLogic {
	return &promoLogic{
		couponDir: couponDir,
		files:     files,
	}
}

type PromoLogic interface {
	ValidatePromo(couponCode string) (bool, error)
}

// ValidatePromo checks if the coupon exists in at least two files
func (l *promoLogic) ValidatePromo(couponCode string) (bool, error) {
	// Channel to collect results from each file
	results := make(chan bool, len(l.files))
	errors := make(chan error, len(l.files))

	var wg sync.WaitGroup

	fmt.Println(l.files)

	// Search each file in a separate goroutine
	for _, filename := range l.files {
		wg.Add(1)
		go func(fname string) {
			defer wg.Done()

			filePath := filepath.Join(l.couponDir, fname)
			file, err := os.Open(filePath)
			if err != nil {
				errors <- err
				return
			}
			defer file.Close()

			// Use buffered reading for memory efficiency
			scanner := bufio.NewScanner(file)
			// Set a larger buffer size for better performance
			buf := make([]byte, 1024*1024)    // 1MB buffer
			scanner.Buffer(buf, 10*1024*1024) // Allow up to 10MB per line

			found := false
			for scanner.Scan() {
				if scanner.Text() == couponCode {
					found = true
					break
				}
			}

			if err := scanner.Err(); err != nil {
				errors <- err
				return
			}

			results <- found
		}(filename)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	foundCount := 0
	for found := range results {
		if found {
			foundCount++
		}
	}

	// Check for any errors
	for err := range errors {
		if err != nil {
			return false, err
		}
	}

	// A coupon is valid only if it exists in at least two files
	return foundCount >= 2, nil
}
