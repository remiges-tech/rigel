package rigel

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/remiges-tech/rigel/etcd"
	"github.com/remiges-tech/rigel/mocks"
	"github.com/remiges-tech/rigel/types"
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
	if schemaFields[0].Name != "key1" || schemaFields[0].Type != "string" {
		t.Errorf("Field 1 is incorrect")
	}
	if schemaFields[1].Name != "key2" || schemaFields[1].Type != "int" {
		t.Errorf("Field 2 is incorrect")
	}
	if schemaFields[2].Name != "key3" || schemaFields[2].Type != "bool" {
		t.Errorf("Field 3 is incorrect")
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
