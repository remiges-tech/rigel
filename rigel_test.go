package rigel

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/remiges-tech/rigel/etcd"
	"github.com/remiges-tech/rigel/mocks"
	"github.com/remiges-tech/rigel/types"
	"github.com/stretchr/testify/assert"
)

func TestNewRigelClient(t *testing.T) {
	etcdStorage := &etcd.EtcdStorage{} // Mocked EtcdStorage
	rigelClient := New(etcdStorage, "app", "module", 1, "config")

	if rigelClient == nil {
		t.Fatalf("Expected rigelClient to be not nil")
	}

	if rigelClient.Storage != etcdStorage {
		t.Errorf("Expected rigelClient.Storage to be equal to provided etcdStorage")
	}
}

func TestGetSchemaFields(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "key1", "type": "string"}, {"name": "key2", "type": "int"}, {"name": "key3", "type": "bool"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Call getSchema
	schemaFields, err := rigelClient.getSchemaFields(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned schema is correct
	if len(schemaFields) != 3 {
		t.Errorf("Returned schema is incorrect")
	}
}

// TestSchemaCache_Population tests that schema is cached after first access
func TestSchemaCache_Population(t *testing.T) {
	getCallCount := 0
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				getCallCount++
				return `[{"name": "key1", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	ctx := context.Background()

	// First call should fetch from storage
	fields1, err := rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected 1 storage call, got %d", getCallCount)
	}

	// Second call should use cache
	fields2, err := rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected still 1 storage call (cached), got %d", getCallCount)
	}

	// Verify both calls returned same data
	if len(fields1) != len(fields2) || fields1[0].Name != fields2[0].Name {
		t.Errorf("Cached data doesn't match original")
	}
}

// TestSchemaCache_DisableCaching tests that schema is fetched from etcd when caching disabled
func TestSchemaCache_DisableCaching(t *testing.T) {
	getCallCount := 0
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				getCallCount++
				return `[{"name": "key1", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	ctx := context.Background()

	// First call with caching enabled
	_, err := rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected 1 storage call, got %d", getCallCount)
	}

	// Disable caching
	rigelClient.DisableCaching()

	// Call should fetch from storage again
	_, err = rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 2 {
		t.Errorf("Expected 2 storage calls (caching disabled), got %d", getCallCount)
	}

	// Another call with caching still disabled
	_, err = rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 3 {
		t.Errorf("Expected 3 storage calls (caching still disabled), got %d", getCallCount)
	}
}

// TestSchemaCache_Invalidation tests that InvalidateSchemaCache works correctly
func TestSchemaCache_Invalidation(t *testing.T) {
	getCallCount := 0
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				getCallCount++
				return `[{"name": "key1", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	ctx := context.Background()

	// First call populates cache
	_, err := rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected 1 storage call, got %d", getCallCount)
	}

	// Second call uses cache
	_, err = rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected still 1 storage call (cached), got %d", getCallCount)
	}

	// Invalidate cache
	rigelClient.InvalidateSchemaCache()

	// Next call should fetch from storage
	_, err = rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 2 {
		t.Errorf("Expected 2 storage calls (cache invalidated), got %d", getCallCount)
	}
}

// TestSchemaCache_ConcurrentAccess verifies that schema caching is thread-safe and efficient
// under concurrent access. This test is critical because:
//
// 1. Multiple goroutines might call getSchemaFields() simultaneously on app startup
// 2. The mutex must properly serialize access to prevent race conditions
// 3. The cache should ensure only ONE etcd call is made, even with 100 concurrent requests
// 4. All goroutines should receive the same correct schema data
//
// This test uses the -race detector when run with `go test -race` to catch data races.
func TestSchemaCache_ConcurrentAccess(t *testing.T) {
	var getCallCount int32
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				atomic.AddInt32(&getCallCount, 1)
				// Simulate etcd latency to increase chance of concurrent access
				time.Sleep(10 * time.Millisecond)
				return `[{"name": "key1", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	ctx := context.Background()

	// Launch 100 goroutines that all try to access schema simultaneously
	var wg sync.WaitGroup
	concurrentRequests := 100
	results := make(chan []types.Field, concurrentRequests)
	
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fields, err := rigelClient.getSchemaFields(ctx)
			if err != nil {
				t.Errorf("Unexpected error in concurrent access: %v", err)
				return
			}
			results <- fields
		}()
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Verify all goroutines got the same correct data
	var firstResult []types.Field
	resultCount := 0
	for fields := range results {
		resultCount++
		if len(fields) != 1 || fields[0].Name != "key1" || fields[0].Type != "string" {
			t.Errorf("Incorrect schema data returned: %+v", fields)
		}
		if firstResult == nil {
			firstResult = fields
		} else if fields[0].Name != firstResult[0].Name {
			t.Errorf("Inconsistent data returned across goroutines")
		}
	}
	
	if resultCount != concurrentRequests {
		t.Errorf("Expected %d results, got %d", concurrentRequests, resultCount)
	}

	// CRITICAL: Verify caching worked - should only call storage ONCE despite 100 requests
	finalCount := atomic.LoadInt32(&getCallCount)
	if finalCount != 1 {
		t.Errorf("Cache failed under concurrency: expected 1 storage call, got %d", finalCount)
	}
	
	// Test concurrent cache control operations don't break the cache
	wg = sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(3)
		// Goroutine 1: Toggles caching
		go func() {
			defer wg.Done()
			rigelClient.DisableCaching()
			rigelClient.EnableCaching()
		}()
		// Goroutine 2: Invalidates cache  
		go func() {
			defer wg.Done()
			rigelClient.InvalidateSchemaCache()
		}()
		// Goroutine 3: Accesses schema
		go func() {
			defer wg.Done()
			rigelClient.getSchemaFields(ctx)
		}()
	}
	wg.Wait()
	
	// Verify system is still functional after concurrent operations
	rigelClient.EnableCaching()
	_, err := rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Errorf("Schema access failed after concurrent operations: %v", err)
	}
}

// TestSchemaCache_AddSchemaInvalidates tests that adding schema clears cache
func TestSchemaCache_AddSchemaInvalidates(t *testing.T) {
	getCallCount := 0
	putCallCount := 0
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				getCallCount++
				return `[{"name": "key1", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
		PutFunc: func(ctx context.Context, key string, value string) error {
			putCallCount++
			return nil
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	ctx := context.Background()

	// First call populates cache
	_, err := rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected 1 storage call, got %d", getCallCount)
	}

	// Second call uses cache
	_, err = rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 1 {
		t.Errorf("Expected still 1 storage call (cached), got %d", getCallCount)
	}

	// Add new schema
	newSchema := types.Schema{
		Version:     1,
		Description: "Updated schema",
		Fields: []types.Field{
			{Name: "key1", Type: "string"},
			{Name: "key2", Type: "int"},
		},
	}
	err = rigelClient.AddSchema(ctx, newSchema)
	if err != nil {
		t.Fatalf("Expected no error adding schema, got %v", err)
	}

	// Next call should fetch from storage (cache invalidated)
	_, err = rigelClient.getSchemaFields(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if getCallCount != 2 {
		t.Errorf("Expected 2 storage calls (cache invalidated by AddSchema), got %d", getCallCount)
	}
}

// BenchmarkSchemaCache_WithCache benchmarks schema access with caching enabled
// 
// Baseline performance (2025-01):
//   BenchmarkSchemaCache_WithCache-24       109828444        10.58 ns/op
//   Hardware: Intel i7-13700K, Linux x86_64
//
// This benchmark should maintain <100 ns/op to ensure cache effectiveness.
func BenchmarkSchemaCache_WithCache(b *testing.B) {
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Simulate etcd latency
			time.Sleep(1 * time.Millisecond)
			return `[{"name": "key1", "type": "string"}]`, nil
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	ctx := context.Background()

	// Pre-populate cache
	rigelClient.getSchemaFields(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rigelClient.getSchemaFields(ctx)
	}
}

// BenchmarkSchemaCache_WithoutCache benchmarks schema access with caching disabled
//
// Baseline performance (2025-01):
//   BenchmarkSchemaCache_WithoutCache-24        1123    1072945 ns/op
//   Hardware: Intel i7-13700K, Linux x86_64
//
// This represents the cost of fetching schema from etcd on every access.
// Cache provides ~101,000x speedup compared to this baseline.
func BenchmarkSchemaCache_WithoutCache(b *testing.B) {
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Simulate etcd latency
			time.Sleep(1 * time.Millisecond)
			return `[{"name": "key1", "type": "string"}]`, nil
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	rigelClient.DisableCaching()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rigelClient.getSchemaFields(ctx)
	}
}

func TestGetSchema(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string for getSchemaFields
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "key1", "type": "string"}, {"name": "key2", "type": "int"}, {"name": "key3", "type": "bool"}]`, nil
			}
			// Return a predefined description for GetSchemaDescriptionPath
			if key == GetSchemaDescriptionPath("app", "module", 1) {
				return "Test schema description", nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Call GetSchema
	schema, err := rigelClient.GetSchema(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned schema is correct
	if len(schema.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(schema.Fields))
	}
	if schema.Description != "Test schema description" {
		t.Errorf("Expected description to be 'Test schema description', got '%s'", schema.Description)
	}
}

func TestGetConfigValue(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined config value JSON string
			if key == GetConfKeyPath("app", "module", 1, "config", "key") {
				return "value", nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Call getConfigValue
	value, err := rigelClient.getConfigValue(context.Background(), "key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned value is correct
	if value != "value" {
		t.Errorf("Expected value to be 'value', got '%v'", value)
	}
}

func TestConstructConfigMap(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string for getSchema
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "key", "type": "string"}]`, nil
			}
			// Return a predefined config value JSON string
			if key == GetConfKeyPath("app", "module", 1, "config", "key") {
				return "value", nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Call constructConfigMap
	configMap, err := rigelClient.constructConfigMap(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned config map is correct
	if value, ok := configMap["key"]; !ok || value != "value" {
		t.Errorf("Expected configMap['key'] to be 'value', got '%v'", value)
	}
}

func TestLoadConfig(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string for getSchema
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "key1", "type": "string"}, {"name": "key2", "type": "int"}, {"name": "key3", "type": "bool"}]`, nil
			}
			// Return a predefined config value JSON string for getConfigValue
			switch key {
			case GetConfKeyPath("app", "module", 1, "config", "key1"):
				return "value1", nil
			case GetConfKeyPath("app", "module", 1, "config", "key2"):
				return `2`, nil
			case GetConfKeyPath("app", "module", 1, "config", "key3"):
				return `true`, nil
			default:
				return "", fmt.Errorf("unexpected key: %s", key)
			}
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	var config struct {
		Key1 string `json:"key1"`
		Key2 int    `json:"key2"`
		Key3 bool   `json:"key3"`
	}
	err := rigelClient.LoadConfig(context.Background(), &config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config.Key1 != "value1" {
		t.Errorf("Expected config.Key1 to be 'value1', got '%s'", config.Key1)
	}
	if config.Key2 != 2 {
		t.Errorf("Expected config.Key2 to be 2, got '%d'", config.Key2)
	}
	if config.Key3 != true {
		t.Errorf("Expected config.Key3 to be true, got '%t'", config.Key3)
	}
}

func TestAddSchema(t *testing.T) {
	// Define schema
	schema := types.Schema{
		Fields: []types.Field{
			{
				Name:        "field1",
				Type:        "string",
				Description: "", // Include empty description to match the actual behavior
				Constraints: &types.Constraints{},
			},
		},
		Description: "description",
		Version:     1,
	}

	// Define expected keys and values
	expectedFieldsKey := GetSchemaFieldsPath("app", "module", 1)
	// Include the description field in the expected JSON
	expectedFieldsValue := `[{"name":"field1","type":"string","description":"","constraints":{}}]`
	expectedDescriptionKey := GetSchemaDescriptionPath("app", "module", 1)
	expectedDescriptionValue := "description"
	// Define the expected key for the field description
	expectedFieldDescriptionKey := expectedFieldsKey + "/field1"

	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		PutFunc: func(ctx context.Context, key string, value string) error {
			switch key {
			case expectedFieldsKey:
				if value != expectedFieldsValue {
					t.Errorf("Expected fields value to be '%s', got '%s'", expectedFieldsValue, value)
				}
			case expectedDescriptionKey:
				if value != expectedDescriptionValue {
					t.Errorf("Expected description value to be '%s', got '%s'", expectedDescriptionValue, value)
				}
			case expectedFieldDescriptionKey:
				// Expect the field description to be stored
				if value != "" {
					t.Errorf("Expected field description value to be empty, got '%s'", value)
				}
			default:
				t.Errorf("Unexpected key: '%s'", key)
			}
			return nil
		},
	}

	rigelClient := NewWithStorage(mockStorage).WithApp("app").WithModule("module")

	// Call AddSchema
	err := rigelClient.AddSchema(context.Background(), schema)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAddSchemaWithTimeout(t *testing.T) {
	// Define schema
	schema := types.Schema{
		Fields: []types.Field{
			{Name: "field1", Type: "string"},
		},
		Description: "description",
	}

	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		PutFunc: func(ctx context.Context, key string, value string) error {
			// Simulate a delay with select to respect context timeout
			select {
			case <-time.After(2 * time.Second):
				return nil // or simulate storage put success
			case <-ctx.Done():
				return ctx.Err() // return the error from the context, which will be a timeout
			}
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Call AddSchema with the context
	err := rigelClient.AddSchema(ctx, schema)
	if err == nil {
		t.Errorf("Expected error due to timeout, got nil")
	}
}

func TestGetInt(t *testing.T) {
	// Create a mock storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined value for a specific key
			switch key {
			case GetConfKeyPath("app", "module", 1, "config", "testParam"):
				return "123", nil
			case GetSchemaFieldsPath("app", "module", 1):
				return `[{"name": "testParam", "type": "int"}]`, nil
			default:
				return "", fmt.Errorf("unexpected key: %s", key)
			}
		},
	}

	// Create a new Rigel instance with a schema
	r := New(mockStorage, "app", "module", 1, "config")
	// Call the GetInt method with a parameter name
	intValue, err := r.GetInt(context.Background(), "testParam")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check if the returned value is correct
	if intValue != 123 {
		t.Errorf("Expected 123, got %d", intValue)
	}
}
func TestGet(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		paramName     string
		expectedValue string
		storageData   map[string]string
		expectError   bool
	}{
		{
			name:          "Get existing value",
			paramName:     "testParam",
			expectedValue: "testValue",
			storageData: map[string]string{
				GetConfKeyPath("app", "module", 1, "config", "testParam"): "testValue",
				GetSchemaFieldsPath("app", "module", 1):                   `[{"name": "testParam", "type": "string"}]`,
			},
			expectError: false,
		},
		{
			name:          "Get non-existing value",
			paramName:     "nonExistingParam",
			expectedValue: "",
			storageData:   map[string]string{},
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock storage
			mockStorage := &mocks.MockStorage{
				GetFunc: func(ctx context.Context, key string) (string, error) {
					value, ok := tc.storageData[key]
					if !ok {
						return "", &KeyNotFoundError{Key: key}
					}
					return value, nil
				},
			}

			// Create a new Rigel instance with a schema
			r := New(mockStorage, "app", "module", 1, "config")

			// Call the Get method with a parameter name
			value, err := r.Get(context.Background(), tc.paramName)

			// Check if the returned value is correct
			if value != tc.expectedValue {
				t.Errorf("Expected '%s', got '%s'", tc.expectedValue, value)
			}

			// Check if error is returned when expected
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got error: %v", tc.expectError, err)
			}
		})
	}
}

func TestGetFromCache(t *testing.T) {
	// Define test case
	paramName := "testParam"
	expectedValue := "testValue"

	// Create a mock cache
	mockCache := &mocks.MockCache{
		GetFunc: func(key string) (string, bool) {
			if key == GetConfKeyPath("app", "module", 1, "config", paramName) {
				return "testValue", true
			}
			return "", false
		},
	}

	// Create a mock storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "testParam", "type": "string"}]`, nil
			}
			t.Errorf("Storage should not be accessed when value is in cache")
			return "", nil
		},
	}

	// Create a new Rigel instance with a schema and cache
	r := New(mockStorage, "app", "module", 1, "config")
	r.Cache = mockCache

	// Call the Get method with a parameter name
	value, err := r.Get(context.Background(), paramName)

	// Check if the returned value is correct
	if value != expectedValue {
		t.Errorf("Expected '%s', got '%s'", expectedValue, value)
	}

	// Check if no error is returned
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestKeyExistsInSchema(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "key1", "type": "string"}, {"name": "key2", "type": "int"}, {"name": "key3", "type": "bool"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Call KeyExistsInSchema
	exists, err := rigelClient.KeyExistsInSchema(context.Background(), "key1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Errorf("Expected key to exist in schema")
	}

	// Call KeyExistsInSchema with a non-existing key
	exists, err = rigelClient.KeyExistsInSchema(context.Background(), "nonExistingKey")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Errorf("Expected key to not exist in schema")
	}
}

func TestSet_KeyNotFoundError(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string with no matching key
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "otherKey", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	// Call Set with a non-existing key
	err := rigelClient.Set(context.Background(), "nonExistingKey", "value")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// Check if the error is a KeyNotFoundError
	if _, ok := err.(*KeyNotFoundError); !ok {
		t.Errorf("Expected error to be a KeyNotFoundError, got %T", err)
	}
}

func TestSet_Success(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string with matching key
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "existingKey", "type": "string"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
		PutFunc: func(ctx context.Context, key string, value string) error {
			// Check if the key and value are correct
			if key != GetConfKeyPath("app", "module", 1, "config", "existingKey") || value != "value" {
				return fmt.Errorf("unexpected key or value: %s, %s", key, value)
			}
			return nil
		},
	}

	// Mocked Cache
	mockCache := &mocks.MockCache{
		SetFunc: func(key string, value string) {
			// Check if the key and value are correct
			if key != GetConfKeyPath("app", "module", 1, "config", "existingKey") || value != "value" {
				t.Errorf("unexpected key or value in cache: %s, %s", key, value)
			}
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	rigelClient.Cache = mockCache

	// Call Set with an existing key
	err := rigelClient.Set(context.Background(), "existingKey", "value")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSet_InvalidJSON(t *testing.T) {
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "config", "type": "json", "description": "JSON config"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
		PutFunc: func(ctx context.Context, key string, value string) error {
			assert.Fail(t, "Put should not be called for invalid JSON")
			return nil
		},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")

	tests := []struct {
		name  string
		value string
	}{
		{"unquoted keys", `{invalid: json}`},
		{"trailing comma", `{"key": "value",}`},
		{"plain text", `not json at all`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rigelClient.Set(context.Background(), "config", tt.value)
			assert.EqualError(t, err, "value does not meet the constraints of the field")
		})
	}
}

func TestSet_ValidJSON(t *testing.T) {
	putCalled := false
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == GetSchemaFieldsPath("app", "module", 1) {
				return `[{"name": "config", "type": "json", "description": "JSON config"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
		PutFunc: func(ctx context.Context, key string, value string) error {
			putCalled = true
			return nil
		},
	}

	mockCache := &mocks.MockCache{
		SetFunc: func(key string, value string) {},
	}

	rigelClient := New(mockStorage, "app", "module", 1, "config")
	rigelClient.Cache = mockCache

	tests := []struct {
		name  string
		value string
	}{
		{"object", `{"key": "value"}`},
		{"array", `[1, 2, 3]`},
		{"nested", `{"users": [{"name": "alice"}], "config": {"timeout": 30}}`},
		{"string", `"hello"`},
		{"number", `42`},
		{"boolean", `true`},
		{"null", `null`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			putCalled = false
			err := rigelClient.Set(context.Background(), "config", tt.value)
			assert.NoError(t, err)
			assert.True(t, putCalled, "Put should be called for valid JSON")
		})
	}
}

func TestGetJSON(t *testing.T) {
	type FeatureFlags struct {
		EnableSignup     bool `json:"enableSignup"`
		EnableOAuth      bool `json:"enableOAuth"`
		MaxLoginAttempts int  `json:"maxLoginAttempts"`
	}

	t.Run("valid JSON unmarshals to struct", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json", "description": "Feature flags"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					return `{"enableSignup": true, "enableOAuth": false, "maxLoginAttempts": 5}`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")

		flags, err := GetJSON[FeatureFlags](rigelClient, context.Background(), "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, true, flags.EnableSignup)
		assert.Equal(t, false, flags.EnableOAuth)
		assert.Equal(t, 5, flags.MaxLoginAttempts)
	})

	t.Run("extra fields in JSON are ignored", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json", "description": "Feature flags"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					return `{"enableSignup": true, "extraField": "ignored", "maxLoginAttempts": 3}`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")

		flags, err := GetJSON[FeatureFlags](rigelClient, context.Background(), "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, true, flags.EnableSignup)
		assert.Equal(t, 3, flags.MaxLoginAttempts)
	})

	t.Run("invalid JSON returns unmarshal error", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json", "description": "Feature flags"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					return `{invalid json}`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")

		_, err := GetJSON[FeatureFlags](rigelClient, context.Background(), "feature.flags")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal JSON")
	})

	t.Run("non-existent key returns error", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "other.key", "type": "json", "description": "Other key"}]`, nil
				}
				return "", &KeyNotFoundError{Key: key}
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")

		_, err := GetJSON[FeatureFlags](rigelClient, context.Background(), "feature.flags")
		assert.Error(t, err)
	})

	t.Run("works with slice type", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "allowed.ips", "type": "json", "description": "Allowed IPs"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "allowed.ips") {
					return `["192.168.1.1", "10.0.0.1", "172.16.0.1"]`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")

		ips, err := GetJSON[[]string](rigelClient, context.Background(), "allowed.ips")
		assert.NoError(t, err)
		assert.Equal(t, []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"}, ips)
	})

	t.Run("works with map type", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "rate.limits", "type": "json", "description": "Rate limits"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "rate.limits") {
					return `{"api": 100, "web": 500, "admin": 1000}`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")

		limits, err := GetJSON[map[string]int](rigelClient, context.Background(), "rate.limits")
		assert.NoError(t, err)
		assert.Equal(t, 100, limits["api"])
		assert.Equal(t, 500, limits["web"])
		assert.Equal(t, 1000, limits["admin"])
	})
}

func ExampleRigel_LoadConfig() {
	//// Create a new EtcdStorage instance
	//etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	//if err != nil {
	//	log.Fatalf("Failed to create EtcdStorage: %v", err)
	//}
	//
	//// Create a new Rigel instance
	//rigelClient := New(etcdStorage)
	//
	//// Define a config struct
	//var config struct {
	//	DatabaseURL string `json:"database_url"`
	//	APIKey      string `json:"api_key"`
	//	IsDebug     bool   `json:"is_debug"`
	//}
	//
	//// Load the config
	//err = rigelClient.LoadConfig("AppConfig", 1, "Production", &config)
	//if err != nil {
	//	log.Fatalf("Failed to load config: %v", err)
	//}
	//
	//// Print the loaded config
	//fmt.Printf("DatabaseURL: %s\n", config.DatabaseURL)
	//fmt.Printf("APIKey: %s\n", config.APIKey)
	//fmt.Printf("IsDebug: %t\n", config.IsDebug)
	//
	//// Output:
	//// DatabaseURL: postgres://user:pass@localhost:5432/dbname
	//// APIKey: abc123
	//// IsDebug: false
}

func TestEnableDisableCaching(t *testing.T) {
	// Create a mock storage
	mockStorage := &mocks.MockStorage{}

	// Create a new Rigel instance (caching is enabled by default)
	r := New(mockStorage, "app", "module", 1, "config")

	// Test that caching is enabled by default
	if !r.IsCachingEnabled() {
		t.Error("Expected caching to be enabled by default")
	}

	// Test DisableCaching
	r.DisableCaching()
	if r.IsCachingEnabled() {
		t.Error("Expected caching to be disabled after DisableCaching()")
	}

	// Test EnableCaching
	r.EnableCaching()
	if !r.IsCachingEnabled() {
		t.Error("Expected caching to be enabled after EnableCaching()")
	}

	// Test method chaining
	r2 := New(mockStorage, "app", "module", 1, "config").DisableCaching()
	if r2.IsCachingEnabled() {
		t.Error("Expected caching to be disabled after chained DisableCaching()")
	}

	r3 := New(mockStorage, "app", "module", 1, "config").DisableCaching().EnableCaching()
	if !r3.IsCachingEnabled() {
		t.Error("Expected caching to be enabled after chained EnableCaching()")
	}
}

func TestIsCachingEnabled(t *testing.T) {
	// Create a mock storage
	mockStorage := &mocks.MockStorage{}

	// Create a new Rigel instance
	r := New(mockStorage, "app", "module", 1, "config")

	// Test IsCachingEnabled with default state
	if !r.IsCachingEnabled() {
		t.Error("Expected IsCachingEnabled() to return true by default")
	}

	// Test after disabling
	r.DisableCaching()
	if r.IsCachingEnabled() {
		t.Error("Expected IsCachingEnabled() to return false after disabling")
	}

	// Test after enabling
	r.EnableCaching()
	if !r.IsCachingEnabled() {
		t.Error("Expected IsCachingEnabled() to return true after enabling")
	}
}

func TestGetWithCachingDisabled(t *testing.T) {
	paramName := "testParam"
	expectedValue := "testValue"
	getCallCount := 0

	// Create a mock storage that counts Get calls
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			switch key {
			case GetSchemaFieldsPath("app", "module", 1):
				return `[{"name": "testParam", "type": "string"}]`, nil
			case GetConfKeyPath("app", "module", 1, "config", paramName):
				getCallCount++
				return expectedValue, nil
			default:
				return "", fmt.Errorf("unexpected key: %s", key)
			}
		},
	}

	// Create a new Rigel instance and disable caching
	r := New(mockStorage, "app", "module", 1, "config").DisableCaching()

	// First Get call
	value1, err := r.Get(context.Background(), paramName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if value1 != expectedValue {
		t.Errorf("Expected '%s', got '%s'", expectedValue, value1)
	}
	if getCallCount != 1 {
		t.Errorf("Expected 1 storage Get call, got %d", getCallCount)
	}

	// Second Get call - should hit storage again since caching is disabled
	value2, err := r.Get(context.Background(), paramName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if value2 != expectedValue {
		t.Errorf("Expected '%s', got '%s'", expectedValue, value2)
	}
	if getCallCount != 2 {
		t.Errorf("Expected 2 storage Get calls, got %d", getCallCount)
	}

	// Verify cache was not used
	_, found := r.Cache.Get(GetConfKeyPath("app", "module", 1, "config", paramName))
	if found {
		t.Error("Expected value not to be in cache when caching is disabled")
	}
}

func TestGetWithCachingEnabled(t *testing.T) {
	paramName := "testParam"
	expectedValue := "testValue"
	getCallCount := 0

	// Create a mock storage that counts Get calls
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			switch key {
			case GetSchemaFieldsPath("app", "module", 1):
				return `[{"name": "testParam", "type": "string"}]`, nil
			case GetConfKeyPath("app", "module", 1, "config", paramName):
				getCallCount++
				return expectedValue, nil
			default:
				return "", fmt.Errorf("unexpected key: %s", key)
			}
		},
	}

	// Create a new Rigel instance (caching enabled by default)
	r := New(mockStorage, "app", "module", 1, "config")

	// First Get call
	value1, err := r.Get(context.Background(), paramName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if value1 != expectedValue {
		t.Errorf("Expected '%s', got '%s'", expectedValue, value1)
	}
	if getCallCount != 1 {
		t.Errorf("Expected 1 storage Get call, got %d", getCallCount)
	}

	// Second Get call - should hit cache
	value2, err := r.Get(context.Background(), paramName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if value2 != expectedValue {
		t.Errorf("Expected '%s', got '%s'", expectedValue, value2)
	}
	if getCallCount != 1 {
		t.Errorf("Expected still 1 storage Get call (cache hit), got %d", getCallCount)
	}
}

func TestWatchConfigWithCachingEnabled(t *testing.T) {
	// Create channels for testing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock storage
	mockStorage := &mocks.MockStorage{
		WatchFunc: func(ctx context.Context, key string, ch chan<- types.Event) error {
			// Simulate events in a goroutine
			go func() {
				// Test case 1: Update existing key in cache
				existingKey := GetConfKeyPath("app", "module", 1, "config", "existing")
				ch <- types.Event{
					Key:   existingKey,
					Value: "newValue",
				}

				// Give time for the first event to be processed
				time.Sleep(20 * time.Millisecond)

				// Test case 2: Add new key (not in cache) - should NOT be added (intentional behavior)
				newKey := GetConfKeyPath("app", "module", 1, "config", "new")
				ch <- types.Event{
					Key:   newKey,
					Value: "brandNewValue",
				}
			}()
			return nil
		},
	}

	// Create a new Rigel instance
	r := New(mockStorage, "app", "module", 1, "config")

	// Pre-populate cache with existing key
	existingKey := GetConfKeyPath("app", "module", 1, "config", "existing")
	r.Cache.Set(existingKey, "oldValue")

	// Start watching
	err := r.WatchConfig(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Give time for the events to be processed
	time.Sleep(50 * time.Millisecond)

	// Check if existing key was updated
	value, found := r.Cache.Get(existingKey)
	if !found {
		t.Error("Expected existing key to still be in cache")
	}
	if value != "newValue" {
		t.Errorf("Expected cache value to be 'newValue', got '%s'", value)
	}

	// Check that new key was NOT added to cache (intentional security behavior)
	newKey := GetConfKeyPath("app", "module", 1, "config", "new")
	_, found = r.Cache.Get(newKey)
	if found {
		t.Error("Expected new key NOT to be added to cache (only existing keys should be updated)")
	}
}

func TestWatchConfigWithCachingDisabled(t *testing.T) {
	// Create channels for testing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock storage
	mockStorage := &mocks.MockStorage{
		WatchFunc: func(ctx context.Context, key string, ch chan<- types.Event) error {
			// Simulate an event in a goroutine
			go func() {
				time.Sleep(10 * time.Millisecond)
				testKey := GetConfKeyPath("app", "module", 1, "config", "test")
				ch <- types.Event{
					Key:   testKey,
					Value: "testValue",
				}
			}()
			return nil
		},
	}

	// Create a new Rigel instance and disable caching
	r := New(mockStorage, "app", "module", 1, "config").DisableCaching()

	// Start watching
	err := r.WatchConfig(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Give time for the event to be processed
	time.Sleep(30 * time.Millisecond)

	// Check that cache was not updated (caching is disabled)
	testKey := GetConfKeyPath("app", "module", 1, "config", "test")
	_, found := r.Cache.Get(testKey)
	if found {
		t.Error("Expected cache not to be updated when caching is disabled")
	}
}

func TestGetJSON_Caching(t *testing.T) {
	type FeatureFlags struct {
		EnableSignup     bool `json:"enableSignup"`
		MaxLoginAttempts int  `json:"maxLoginAttempts"`
	}

	t.Run("returns cached parsed object when value unchanged", func(t *testing.T) {
		parseCount := 0
		jsonValue := `{"enableSignup": true, "maxLoginAttempts": 5}`

		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					return jsonValue, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")
		ctx := context.Background()

		// First call - should parse
		flags1, err := GetJSON[FeatureFlags](rigelClient, ctx, "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, true, flags1.EnableSignup)

		// Second call - should return cached
		flags2, err := GetJSON[FeatureFlags](rigelClient, ctx, "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, true, flags2.EnableSignup)

		// Verify we got cached result by checking internal state
		rigelClient.parsedCacheMu.RLock()
		cacheKey := GetConfKeyPath("app", "module", 1, "config", "feature.flags")
		entry, ok := rigelClient.parsedCache[cacheKey]
		rigelClient.parsedCacheMu.RUnlock()

		assert.True(t, ok, "parsed cache should have entry")
		assert.Equal(t, jsonValue, entry.rawValue)
		_ = parseCount // parseCount tracking would require instrumenting json.Unmarshal
	})

	t.Run("re-parses when underlying value changes", func(t *testing.T) {
		callCount := 0
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					callCount++
					if callCount == 1 {
						return `{"enableSignup": true, "maxLoginAttempts": 5}`, nil
					}
					return `{"enableSignup": false, "maxLoginAttempts": 10}`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")
		ctx := context.Background()

		// First call
		flags1, err := GetJSON[FeatureFlags](rigelClient, ctx, "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, true, flags1.EnableSignup)
		assert.Equal(t, 5, flags1.MaxLoginAttempts)

		// Simulate cache miss by clearing string cache
		rigelClient.Cache.Delete(GetConfKeyPath("app", "module", 1, "config", "feature.flags"))

		// Second call - value changed
		flags2, err := GetJSON[FeatureFlags](rigelClient, ctx, "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, false, flags2.EnableSignup)
		assert.Equal(t, 10, flags2.MaxLoginAttempts)
	})

	t.Run("caches parsed objects even with DisableCaching", func(t *testing.T) {
		jsonValue := `{"enableSignup": true, "maxLoginAttempts": 5}`
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					return jsonValue, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config").DisableCaching()
		ctx := context.Background()

		// Call GetJSON
		flags, err := GetJSON[FeatureFlags](rigelClient, ctx, "feature.flags")
		assert.NoError(t, err)
		assert.Equal(t, true, flags.EnableSignup)

		// Parsed cache should still be populated (independent of string cache setting)
		rigelClient.parsedCacheMu.RLock()
		cacheKey := GetConfKeyPath("app", "module", 1, "config", "feature.flags")
		entry, ok := rigelClient.parsedCache[cacheKey]
		rigelClient.parsedCacheMu.RUnlock()

		assert.True(t, ok, "parsed cache should have entry even with DisableCaching")
		assert.Equal(t, jsonValue, entry.rawValue)
	})

	t.Run("concurrent access is safe", func(t *testing.T) {
		mockStorage := &mocks.MockStorage{
			GetFunc: func(ctx context.Context, key string) (string, error) {
				if key == GetSchemaFieldsPath("app", "module", 1) {
					return `[{"name": "feature.flags", "type": "json"}]`, nil
				}
				if key == GetConfKeyPath("app", "module", 1, "config", "feature.flags") {
					time.Sleep(5 * time.Millisecond)
					return `{"enableSignup": true, "maxLoginAttempts": 5}`, nil
				}
				return "", fmt.Errorf("unexpected key: %s", key)
			},
		}

		rigelClient := New(mockStorage, "app", "module", 1, "config")
		ctx := context.Background()

		var wg sync.WaitGroup
		concurrentRequests := 50
		results := make(chan FeatureFlags, concurrentRequests)
		errors := make(chan error, concurrentRequests)

		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				flags, err := GetJSON[FeatureFlags](rigelClient, ctx, "feature.flags")
				if err != nil {
					errors <- err
					return
				}
				results <- flags
			}()
		}

		wg.Wait()
		close(results)
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Unexpected error in concurrent access: %v", err)
		}

		// Verify all results are correct
		resultCount := 0
		for flags := range results {
			resultCount++
			assert.Equal(t, true, flags.EnableSignup)
			assert.Equal(t, 5, flags.MaxLoginAttempts)
		}
		assert.Equal(t, concurrentRequests, resultCount)
	})
}

func TestConvertToType(t *testing.T) {
	tests := []struct {
		name      string
		valueStr  string
		fieldType string
		want      interface{}
		wantErr   bool
	}{
		{
			name:      "valid json object",
			valueStr:  `{"key": "value"}`,
			fieldType: "json",
			want:      `{"key": "value"}`,
			wantErr:   false,
		},
		{
			name:      "valid json array",
			valueStr:  `[1, 2, 3]`,
			fieldType: "json",
			want:      `[1, 2, 3]`,
			wantErr:   false,
		},
		{
			name:      "valid nested json with object and array",
			valueStr:  `{"users": [{"name": "alice", "roles": ["admin", "user"]}, {"name": "bob", "roles": ["user"]}], "config": {"timeout": 30, "retries": 3}}`,
			fieldType: "json",
			want:      `{"users": [{"name": "alice", "roles": ["admin", "user"]}, {"name": "bob", "roles": ["user"]}], "config": {"timeout": 30, "retries": 3}}`,
			wantErr:   false,
		},
		{
			name:      "valid json string",
			valueStr:  `"hello"`,
			fieldType: "json",
			want:      `"hello"`,
			wantErr:   false,
		},
		{
			name:      "valid json number",
			valueStr:  `42`,
			fieldType: "json",
			want:      `42`,
			wantErr:   false,
		},
		{
			name:      "valid json boolean",
			valueStr:  `true`,
			fieldType: "json",
			want:      `true`,
			wantErr:   false,
		},
		{
			name:      "valid json null",
			valueStr:  `null`,
			fieldType: "json",
			want:      `null`,
			wantErr:   false,
		},
		{
			name:      "invalid json - missing quotes",
			valueStr:  `{key: value}`,
			fieldType: "json",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid json - trailing comma",
			valueStr:  `{"key": "value",}`,
			fieldType: "json",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid json - plain text",
			valueStr:  `not json at all`,
			fieldType: "json",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "int type still works",
			valueStr:  "123",
			fieldType: "int",
			want:      123,
			wantErr:   false,
		},
		{
			name:      "bool type still works",
			valueStr:  "true",
			fieldType: "bool",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "float type still works",
			valueStr:  "3.14",
			fieldType: "float",
			want:      3.14,
			wantErr:   false,
		},
		{
			name:      "string type still works",
			valueStr:  "hello",
			fieldType: "string",
			want:      "hello",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToType(tt.valueStr, tt.fieldType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
