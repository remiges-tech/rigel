package mocks

import (
	"context"
	"fmt"
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
		fmt.Printf("Expected no error, got %v\n", err)
		return
	}

	// Check the returned value
	if value != "testValue" {
		fmt.Printf("Expected value to be 'testValue', got '%s'\n", value)
		return
	}

	// Print the value
	fmt.Println(value)
	// Output: testValue
}
