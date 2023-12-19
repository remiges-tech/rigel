package rigel

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/remiges-tech/rigel/etcd"
	"github.com/remiges-tech/rigel/types"
)

const (
	rigelPrefix          = "/remiges/rigel"
	schemaDescriptionKey = "description"
	schemaNameKey        = "name"
	schemaVersionKey     = "version"
	schemaFieldsKey      = "fields"
	defaultEtcdEndpoints = "localhost:2379"
)

// Rigel represents a client for Rigel configuration manager server.
type Rigel struct {
	Storage       types.Storage
	schemaName    string
	schemaVersion int
	configName    string
	Cache         types.Cache
}

// WithConfig sets the schemaName, schemaVersion, and configName for the Rigel instance and returns a new instance.
func (r *Rigel) WithConfig(schemaName string, schemaVersion int, configName string) *Rigel {
	return &Rigel{
		Storage:       r.Storage,
		schemaName:    schemaName,
		schemaVersion: schemaVersion,
		configName:    configName,
		Cache:         r.Cache,
	}
}

// New creates a new instance of Rigel with the provided Storage interface.
// The Storage interface is used by Rigel to interact with the underlying storage system.
// Currently, only etcd is supported as a storage system.
func New(storage types.Storage) *Rigel {
	return &Rigel{
		Storage: storage,
		Cache:   NewInMemoryCache(),
	}
}

// Default creates a new instance of Rigel with a default EtcdStorage instance.
func Default() (*Rigel, error) {
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		return nil, fmt.Errorf("failed to create default EtcdStorage: %w", err)
	}

	return &Rigel{
		Storage: etcdStorage,
	}, nil
}

// LoadConfig retrieves the configuration data associated with the provided configName.
// It then unmarshals this data into the provided configStruct.
//
// The configStruct parameter must be a pointer to a config struct used in the application.
// If it is not, an error will be returned.
// Non-pointer or non-struct types aren't supported due to type safety issues (e.g., unexpected fields in JSON)
// and modification restrictions, as non-pointer variables can't be updated by json.Unmarshal.
func (r *Rigel) LoadConfig(ctx context.Context, schemaName string, schemaVersion int, configName string, configStruct any) error {
	// Check if configStruct is a pointer to a struct
	val := reflect.ValueOf(configStruct)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("configStruct must be a pointer to a struct")
	}

	// Construct the configuration map
	configMap, err := r.constructConfigMap(ctx, schemaName, schemaVersion)
	if err != nil {
		return err
	}

	// Marshal the configuration map into a JSON string
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Unmarshal the JSON string into the provided configStruct
	err = json.Unmarshal(configJSON, configStruct)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config value: %w", err)
	}

	return nil
}

// AddSchema adds a new schema to the Rigel storage.
// If a schema with the same name and version already exists in the storage,
// AddSchema will override the existing schema with the new one.
func (r *Rigel) AddSchema(ctx context.Context, schema types.Schema) error {
	// Convert fields to JSON
	fieldsJson, err := json.Marshal(schema.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %v", err)
	}

	// Get the base schema path
	baseSchemaPath := getSchemaPath(schema.Name, schema.Version)

	// Store fields
	fieldsKey := baseSchemaPath + schemaFieldsKey
	err = r.Storage.Put(ctx, fieldsKey, string(fieldsJson))
	if err != nil {
		return fmt.Errorf("failed to store fields: %v", err)
	}

	// Store description
	descriptionKey := baseSchemaPath + schemaDescriptionKey
	err = r.Storage.Put(ctx, descriptionKey, schema.Description)
	if err != nil {
		return fmt.Errorf("failed to store description: %v", err)
	}

	// Store name
	nameKey := baseSchemaPath + schemaNameKey
	err = r.Storage.Put(ctx, nameKey, schema.Name)
	if err != nil {
		return fmt.Errorf("failed to store name: %v", err)
	}

	// Store version
	versionKey := baseSchemaPath + schemaVersionKey
	err = r.Storage.Put(ctx, versionKey, strconv.Itoa(schema.Version))
	if err != nil {
		return fmt.Errorf("failed to store version: %v", err)
	}

	return nil
}

// getSchema retrieves a schema from the storage based on the provided schemaName and schemaVersion.
func (r *Rigel) getSchema(ctx context.Context, schemaName string, schemaVersion int) (*types.Schema, error) {
	// Construct the base key for the schema
	schemaFieldsKey := getSchemaFieldsPath(schemaName, schemaVersion)

	fieldsStr, err := r.Storage.Get(ctx, schemaFieldsKey)
	if err != nil {
		return nil, err
	}
	var fields []types.Field
	err = json.Unmarshal([]byte(fieldsStr), &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
	}

	// Construct the schema
	schema := &types.Schema{
		Name:    schemaName,
		Version: schemaVersion,
		Fields:  fields,
	}

	return schema, nil
}

// getConfigValue retrieves a configuration value from Rigel based on the provided schemaName, schemaVersion, and paramName.
func (r *Rigel) getConfigValue(ctx context.Context, schemaName string, schemaVersion int, paramName string) (string, error) {
	// Construct the key for the parameter
	key := getConfKeyPath(schemaName, schemaVersion, paramName)

	// Retrieve the parameter value from the storage
	value, err := r.Storage.Get(ctx, key)
	if err != nil {
		return "", err
	}

	return value, nil
}

// constructConfigMap constructs a configuration map based on the provided schema, schemaName, and schemaVersion.
func (r *Rigel) constructConfigMap(ctx context.Context, schemaName string, schemaVersion int) (map[string]any, error) {
	// Retrieve the schema
	schema, err := r.getSchema(ctx, schemaName, schemaVersion)
	if err != nil {
		return nil, err
	}
	// Construct the configuration map
	config := make(map[string]any)
	for _, field := range schema.Fields {
		// Retrieve the configuration value for the field
		valueStr, err := r.getConfigValue(ctx, schemaName, schemaVersion, field.Name)
		if err != nil {
			return nil, err
		}

		// Convert the value to the correct type based on the field type
		// In constructConfigMap:
		value, err := convertToType(valueStr, field.Type)
		if err != nil {
			return nil, err
		}

		// Add the value to the configuration map
		config[field.Name] = value
	}
	return config, nil
}

// getFieldType retrieves the type of the field specified by paramName from the schema.
func (r *Rigel) getFieldType(ctx context.Context, paramName string) (string, error) {
	// Retrieve the schema
	schema, err := r.getSchema(ctx, r.schemaName, r.schemaVersion)
	if err != nil {
		return "", err
	}

	// Find the type of the field specified by paramName
	for _, field := range schema.Fields {
		if field.Name == paramName {
			return field.Type, nil
		}
	}

	return "", fmt.Errorf("field %s not found in schema", paramName)
}

// get retrieves a value from the storage based on the provided key.
// It converts the retrieved value to the correct type based on the field type.
// If the field type is not "int" or "bool", the value is assumed to be a string.
// get retrieves a value from the cache or storage and returns it as a string.
func (r *Rigel) get(ctx context.Context, paramName string) (string, error) {
	// Construct the key for the parameter
	key := getConfKeyPath(r.schemaName, r.schemaVersion, paramName)

	// Try to get the value from the cache
	value, found := r.Cache.Get(key)
	if found {
		return value.(string), nil
	}

	// If the value is not in the cache, retrieve it from the storage
	valueStr, err := r.Storage.Get(ctx, key)
	if err != nil {
		return "", err
	}

	// Store the value in the cache
	r.Cache.Set(key, valueStr)

	return valueStr, nil
}

func (r *Rigel) GetInt(ctx context.Context, paramName string) (int, error) {
	valueStr, err := r.get(ctx, paramName)
	if err != nil {
		return 0, err
	}
	intValue, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to int: %w", err)
	}
	return intValue, nil
}

func (r *Rigel) GetBool(ctx context.Context, paramName string) (bool, error) {
	valueStr, err := r.get(ctx, paramName)
	if err != nil {
		return false, err
	}
	boolValue, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert value to bool: %w", err)
	}
	return boolValue, nil
}

func (r *Rigel) GetString(ctx context.Context, paramName string) (string, error) {
	valueStr, err := r.get(ctx, paramName)
	if err != nil {
		return "", err
	}
	return valueStr, nil
}

// convertToType converts a string value to the specified type.
func convertToType(valueStr string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "int":
		intValue, err := strconv.Atoi(valueStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int: %w", err)
		}
		return intValue, nil
	case "bool":
		boolValue, err := strconv.ParseBool(valueStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to bool: %w", err)
		}
		return boolValue, nil
	default: // "string"
		return valueStr, nil
	}
}
