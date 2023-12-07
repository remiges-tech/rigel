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
	Storage types.Storage
}

// New creates a new instance of Rigel with the provided Storage interface.
// The Storage interface is used by Rigel to interact with the underlying storage system.
// Currently, only etcd is supported as a storage system.
func New(storage types.Storage) *Rigel {
	return &Rigel{
		Storage: storage,
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
		var value any
		switch field.Type {
		case "int":
			value, err = strconv.Atoi(valueStr)
			if err != nil {
				return nil, fmt.Errorf("failed to convert value to int: %w", err)
			}
		case "bool":
			value, err = strconv.ParseBool(valueStr)
			if err != nil {
				return nil, fmt.Errorf("failed to convert value to bool: %w", err)
			}
		default:
			// Assume the value is a string if the field type is not "int" or "bool"
			value = valueStr
		}

		// Add the value to the configuration map
		config[field.Name] = value
	}
	return config, nil
}
