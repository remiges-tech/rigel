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
	rigelClient := New(etcdStorage)

	if rigelClient == nil {
		t.Fatalf("Expected rigelClient to be not nil")
	}

	if rigelClient.Storage != etcdStorage {
		t.Errorf("Expected rigelClient.Storage to be equal to provided etcdStorage")
	}
}

func TestGetSchema(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined schema JSON string
			if key == getSchemaFieldsPath("schemaName", 1) {
				return `[{"name": "key1", "type": "string"}, {"name": "key2", "type": "int"}, {"name": "key3", "type": "bool"}]`, nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	rigelClient := New(mockStorage)

	// Call getSchema
	schema, err := rigelClient.getSchema(context.Background(), "schemaName", 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned schema is correct
	if schema.Name != "schemaName" || schema.Version != 1 || len(schema.Fields) != 3 {
		t.Errorf("Returned schema is incorrect")
	}
	if schema.Fields[0].Name != "key1" || schema.Fields[0].Type != "string" {
		t.Errorf("Field 1 is incorrect")
	}
	if schema.Fields[1].Name != "key2" || schema.Fields[1].Type != "int" {
		t.Errorf("Field 2 is incorrect")
	}
	if schema.Fields[2].Name != "key3" || schema.Fields[2].Type != "bool" {
		t.Errorf("Field 3 is incorrect")
	}
}

func TestGetConfigValue(t *testing.T) {
	// Mocked Storage
	mockStorage := &mocks.MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			// Return a predefined config value JSON string
			return "value", nil
		},
	}

	rigelClient := New(mockStorage)

	// Call getConfigValue
	value, err := rigelClient.getConfigValue(context.Background(), "schemaName", 1, "key")
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
			if key == getSchemaFieldsPath("schemaName", 1) {
				return `[{"name": "key", "type": "string"}]`, nil
			}
			// Return a predefined config value JSON string
			return "value", nil
		},
	}

	rigelClient := New(mockStorage)

	// Call constructConfigMap
	configMap, err := rigelClient.constructConfigMap(context.Background(), "schemaName", 1)
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
			if key == getSchemaFieldsPath("schemaName", 1) {
				return `[{"name": "key1", "type": "string"}, {"name": "key2", "type": "int"}, {"name": "key3", "type": "bool"}]`, nil
			}
			// Return a predefined config value JSON string for getConfigValue
			switch key {
			case getConfKeyPath("schemaName", 1, "key1"):
				return "value1", nil
			case getConfKeyPath("schemaName", 1, "key2"):
				return `2`, nil
			case getConfKeyPath("schemaName", 1, "key3"):
				return `true`, nil
			default:
				return "", fmt.Errorf("unexpected key: %s", key)
			}
		},
	}

	rigelClient := New(mockStorage)

	var config struct {
		Key1 string `json:"key1"`
		Key2 int    `json:"key2"`
		Key3 bool   `json:"key3"`
	}
	err := rigelClient.LoadConfig(context.Background(), "schemaName", 1, "configName", &config)
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
		Name:        "schemaName",
		Version:     1,
		Description: "description",
		Fields: []types.Field{
			{Name: "field1", Type: "string"},
		},
	}

	// Define expected keys and values
	expectedFieldsKey := getSchemaPath("schemaName", 1) + schemaFieldsKey
	expectedFieldsValue := `[{"name":"field1","type":"string"}]`
	expectedDescriptionKey := getSchemaPath("schemaName", 1) + schemaDescriptionKey
	expectedDescriptionValue := "description"
	expectedNameKey := getSchemaPath("schemaName", 1) + schemaNameKey
	expectedNameValue := "schemaName"
	expectedVersionKey := getSchemaPath("schemaName", 1) + schemaVersionKey
	expectedVersionValue := "1"

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
			case expectedNameKey:
				if value != expectedNameValue {
					t.Errorf("Expected name value to be '%s', got '%s'", expectedNameValue, value)
				}
			case expectedVersionKey:
				if value != expectedVersionValue {
					t.Errorf("Expected version value to be '%s', got '%s'", expectedVersionValue, value)
				}
			default:
				t.Errorf("Unexpected key: '%s'", key)
			}
			return nil
		},
	}

	rigelClient := New(mockStorage)

	// Call AddSchema
	err := rigelClient.AddSchema(context.Background(), schema)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAddSchemaWithTimeout(t *testing.T) {
	// Define schema
	schema := types.Schema{
		Name:        "schemaName",
		Version:     1,
		Description: "description",
		Fields: []types.Field{
			{Name: "field1", Type: "string"},
		},
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

	rigelClient := New(mockStorage)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Call AddSchema with the context
	err := rigelClient.AddSchema(ctx, schema)
	if err == nil {
		t.Errorf("Expected error due to timeout, got nil")
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
