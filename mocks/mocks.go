// Package mocks provides mock implementations of the interfaces used in Rigel.
// These mocks can be used in tests to simulate the behavior of the real implementations.
package mocks

import (
	"context"
)

// MockStorage is a mock implementation of the Storage interface.
// It's used for testing purposes to simulate the behavior of a real Storage implementation.
// Each method of the Storage interface is represented as a function field in this struct.
// These function fields can be set to specific functions in tests to control the mock's behavior.
type MockStorage struct {
	GetFunc func(ctx context.Context, key string) (string, error)
	PutFunc func(ctx context.Context, key string, value string) error
}

// Get is a method that implements the Get method of the Storage interface.
// It returns the function stored in the GetFunc field of the MockStorage struct
// and returns the result.
func (m *MockStorage) Get(ctx context.Context, key string) (string, error) {
	return m.GetFunc(ctx, key)
}

// Put is a method that implements the Put method of the Storage interface.
// It calls the function stored in the PutFunc field of the MockStorage struct
// and returns the result.
func (m *MockStorage) Put(ctx context.Context, key string, value string) error {
	return m.PutFunc(ctx, key, value)
}
