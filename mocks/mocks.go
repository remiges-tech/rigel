// Package mocks provides mock implementations of the interfaces used in Rigel.
// These mocks can be used in tests to simulate the behavior of the real implementations.
package mocks

import (
	"context"
	"sync"

	"github.com/remiges-tech/rigel/types"
)

// MockStorage is a mock implementation of the Storage interface.
// It's used for testing purposes to simulate the behavior of a real Storage implementation.
// Each method of the Storage interface is represented as a function field in this struct.
// These function fields can be set to specific functions in tests to control the mock's behavior.
type MockStorage struct {
	GetFunc   func(ctx context.Context, key string) (string, error)
	PutFunc   func(ctx context.Context, key string, value string) error
	WatchFunc func(ctx context.Context, key string, ch chan<- types.Event) error
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

// MockCache is a mock implementation of the Cache interface.
// It's used for testing purposes to simulate the behavior of a real Cache implementation.
// Each method of the Cache interface is represented as a function field in this struct.
// These function fields can be set to specific functions in tests to control the mock's behavior.
type MockCache struct {
	data       map[string]string
	mu         sync.RWMutex
	GetFunc    func(key string) (string, bool)
	SetFunc    func(key string, value string)
	DeleteFunc func(key string)
}

// Get is a method that implements the Get method of the Cache interface.
// It returns the function stored in the GetFunc field of the MockCache struct
// and returns the result.
func (c *MockCache) Get(key string) (string, bool) {
	if c.GetFunc != nil {
		return c.GetFunc(key)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, found := c.data[key]
	return value, found
}

// Set is a method that implements the Set method of the Cache interface.
// It calls the function stored in the SetFunc field of the MockCache struct.
func (c *MockCache) Set(key string, value string) {
	if c.SetFunc != nil {
		c.SetFunc(key, value)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Delete is a method that implements the Delete method of the Cache interface.
// It calls the function stored in the DeleteFunc field of the MockCache struct.
func (c *MockCache) Delete(key string) {
	if c.DeleteFunc != nil {
		c.DeleteFunc(key)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Watch is a method that implements the Watch method of the Storage interface.
// It calls the function stored in the WatchFunc field of the MockStorage struct
// and returns the result.
func (m *MockStorage) Watch(ctx context.Context, key string, ch chan<- types.Event) error {
	return m.WatchFunc(ctx, key, ch)
}
