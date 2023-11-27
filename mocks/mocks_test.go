package mocks

import (
	"context"
	"fmt"
	"log"
)

func ExampleMockStorage_Get() {
	// Create a new MockStorage instance
	mockStorage := &MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == "testKey" {
				return "testValue", nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	// Call the Get method with "testKey"
	value, err := mockStorage.Get(context.Background(), "testKey")
	if err != nil {
		log.Fatalf("Expected no error, got %v", err)
	}

	// Check the returned value
	if value != "testValue" {
		log.Fatalf("Expected value to be 'testValue', got '%s'", value)
	}
	// Output: testValue
}
