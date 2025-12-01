package rigel

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

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
	Storage      types.Storage
	Cache        types.Cache
	App          string
	Module       string
	Version      int
	Config       string
	cacheEnabled atomic.Bool
	mu           sync.Mutex
	schemaCache  []types.Field
	schemaLoaded atomic.Bool
}

// New creates a new instance of Rigel with the provided Storage interface.
// The Storage interface is used by Rigel to interact with the underlying storage system.
// Currently, only etcd is supported as a storage system.
// By default, caching is enabled.
func New(storage types.Storage, app string, module string, version int, config string) *Rigel {
	r := &Rigel{
		Storage: storage,
		Cache:   NewInMemoryCache(),
		App:     app,
		Module:  module,
		Version: version,
		Config:  config,
	}
	r.cacheEnabled.Store(true)
	return r
}

// NewWithStorage creates a new instance of Rigel with the provided Storage interface.
// This function is useful when you want to create a Rigel object with a specific storage system,
// but you don't want to set the other parameters (app, module, version, config) at the time of creation.
// This is typically used in admin tasks like schema creation where version field is not known while
// adding a new schema definition. Once Rigel object is constructed using NewWithStorage other
// required params for admin tasks are supposed to be added using the with-prefixed functions like
// WithApp, WithModule, etc.
// By default, caching is enabled.
func NewWithStorage(storage types.Storage) *Rigel {
	r := &Rigel{
		Storage: storage,
		Cache:   NewInMemoryCache(),
	}
	r.cacheEnabled.Store(true)
	return r
}

// WithApp sets the App field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithApp(app string) *Rigel {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.App = app
	return r
}

// WithModule sets the Module field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithModule(module string) *Rigel {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Module = module
	return r
}

// WithVersion sets the Version field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithVersion(version int) *Rigel {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Version = version
	return r
}

// WithConfig sets the Config field of the Rigel struct and returns the modified Rigel object.
// This method is typically used for method chaining during Rigel object creation.
func (r *Rigel) WithConfig(config string) *Rigel {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Config = config
	return r
}

// EnableCaching enables the use of the cache for Get operations and etcd watches.
// When enabled, values are cached locally for faster access, and the cache is updated
// when changes occur in etcd (if WatchConfig is called).
func (r *Rigel) EnableCaching() *Rigel {
	r.cacheEnabled.Store(true)
	return r
}

// DisableCaching disables the use of the cache for Get operations.
// When disabled, all Get operations fetch values directly from etcd,
// bypassing the local cache entirely. This ensures values are always fresh
// from etcd but may impact performance with increased network calls.
// This also clears the schema cache to ensure fresh schema lookups.
func (r *Rigel) DisableCaching() *Rigel {
	r.cacheEnabled.Store(false)
	// Clear schema cache flag atomically
	r.schemaLoaded.Store(false)
	return r
}

// IsCachingEnabled returns whether caching is currently enabled for this Rigel instance.
func (r *Rigel) IsCachingEnabled() bool {
	return r.cacheEnabled.Load()
}

// InvalidateSchemaCache clears the cached schema fields, forcing a fresh
// load from storage on the next access. This is useful when schema changes
// are made and need to be reflected immediately.
func (r *Rigel) InvalidateSchemaCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schemaLoaded.Store(false)
	r.schemaCache = nil
}

// Default creates a new instance of Rigel with a default EtcdStorage instance.
// By default, caching is enabled.
func Default() (*Rigel, error) {
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		return nil, fmt.Errorf("failed to create default EtcdStorage: %w", err)
	}

	r := &Rigel{
		Storage: etcdStorage,
		Cache:   NewInMemoryCache(),
	}
	r.cacheEnabled.Store(true)
	return r, nil
}

// KeyExistsInSchema checks if a key exists in the schema.
func (r *Rigel) KeyExistsInSchema(ctx context.Context, key string) (bool, error) {
	schemaFields, err := r.getSchemaFields(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get schema: %w", err)
	}

	for _, field := range schemaFields {
		if field.Name == key {
			return true, nil
		}
	}

	return false, nil
}

// Set sets a value of a config key in the storage.
func (r *Rigel) Set(ctx context.Context, configKey string, value string) error {
	// Check if the key exists in the schema
	exists, err := r.KeyExistsInSchema(ctx, configKey)
	if err != nil {
		return fmt.Errorf("failed to check if key exists in schema: %w", err)
	}
	if !exists {
		return &KeyNotFoundError{Key: configKey}
	}

	// Get the schema
	schemaFields, err := r.getSchemaFields(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	// Find the field in the schema
	var field *types.Field
	for _, f := range schemaFields {
		if f.Name == configKey {
			field = &f
			break
		}
	}

	// Validate the value against the field's constraints
	if !ValidateValueAgainstConstraints(value, field) {
		return fmt.Errorf("value does not meet the constraints of the field")
	}

	// Construct the key for the parameter
	key := GetConfKeyPath(r.App, r.Module, r.Version, r.Config, configKey)

	// Set the value in the storage
	err = r.Storage.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	// Update the value in the cache
	r.Cache.Set(key, value)

	return nil
}

// LoadConfig retrieves the configuration data associated with the provided configName.
// It then unmarshals this data into the provided configStruct.
//
// The configStruct parameter must be a pointer to a config struct used in the application.
// If it is not, an error will be returned.
// Non-pointer or non-struct types aren't supported due to type safety issues (e.g., unexpected fields in JSON)
// and modification restrictions, as non-pointer variables can't be updated by json.Unmarshal.
func (r *Rigel) LoadConfig(ctx context.Context, configStruct any) error {
	// Check if configStruct is a pointer to a struct
	val := reflect.ValueOf(configStruct)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("configStruct must be a pointer to a struct")
	}

	// Construct the configuration map
	configMap, err := r.constructConfigMap(ctx)
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

	// Get the base schema path using the version from the schema
	baseSchemaPath := GetSchemaPath(r.App, r.Module, schema.Version)

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

	for _, field := range schema.Fields {
		//store felid description
		felidDescriptionKey := baseSchemaPath + schemaFieldsKey + "/" + field.Name
		err = r.Storage.Put(ctx, felidDescriptionKey, field.Description)
		if err != nil {
			return fmt.Errorf("failed to store description: %v", err)
		}
	}

	// Invalidate schema cache when a new schema is added
	r.InvalidateSchemaCache()

	return nil
}

// getSchemaFields retrieves the schema fields
func (r *Rigel) getSchemaFields(ctx context.Context) ([]types.Field, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check schema cache first if caching is enabled
	// Note: We check cacheEnabled first to ensure DisableCaching() bypasses schema cache
	if r.cacheEnabled.Load() && r.schemaLoaded.Load() {
		return r.schemaCache, nil
	}

	// Load from etcd
	schemaFieldsKey := GetSchemaFieldsPath(r.App, r.Module, r.Version)

	fieldsStr, err := r.Storage.Get(ctx, schemaFieldsKey)
	if err != nil {
		return nil, err
	}

	var fields []types.Field
	err = json.Unmarshal([]byte(fieldsStr), &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
	}

	// Cache the schema if caching is enabled
	if r.cacheEnabled.Load() {
		r.schemaCache = fields
		r.schemaLoaded.Store(true)
	}

	return fields, nil
}

// GetSchema retrieves a schema (fields and metadata)
func (r *Rigel) GetSchema(ctx context.Context) (*types.Schema, error) {
	schemaDescriptionKey := GetSchemaDescriptionPath(r.App, r.Module, r.Version)
	description, err := r.Storage.Get(ctx, schemaDescriptionKey)
	if err != nil {
		return nil, err
	}

	fields, err := r.getSchemaFields(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema fields: %w", err)
	}

	schema := &types.Schema{
		Version:     r.Version,
		Fields:      fields,
		Description: description,
	}
	return schema, nil
}

// getConfigValue retrieves a configuration value from Rigel based on the provided schemaName, schemaVersion, and paramName.
func (r *Rigel) getConfigValue(ctx context.Context, paramName string) (string, error) {
	// Construct the key for the parameter
	key := GetConfKeyPath(r.App, r.Module, r.Version, r.Config, paramName)

	// Retrieve the parameter value from the storage
	value, err := r.Storage.Get(ctx, key)
	if err != nil {
		return "", err
	}

	return value, nil
}

// constructConfigMap constructs a configuration map based on the provided schema, schemaName, and schemaVersion.
// constructConfigMap constructs a configuration map based on the Rigel object.
func (r *Rigel) constructConfigMap(ctx context.Context) (map[string]any, error) {
	// Retrieve the schema
	schemaFields, err := r.getSchemaFields(ctx)
	if err != nil {
		return nil, err
	}
	// Construct the configuration map
	config := make(map[string]any)
	for _, field := range schemaFields {
		// Retrieve the configuration value for the field
		valueStr, err := r.getConfigValue(ctx, field.Name)
		if err != nil {
			return nil, err
		}

		// Convert the value to the correct type based on the field type
		value, err := convertToType(valueStr, field.Type)
		if err != nil {
			return nil, err
		}

		// Add the value to the configuration map
		config[field.Name] = value
	}
	return config, nil
}

type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key %s not found in config", e.Key)
}

// Get retrieves a value from the storage based on the provided key.
// It converts the retrieved value to the correct type based on the field type.
// If the field type is not "int" or "bool", the value is assumed to be a string.
// If caching is enabled, it tries to get the value from the cache first, otherwise
// it retrieves it directly from the storage.
func (r *Rigel) Get(ctx context.Context, configKey string) (string, error) {
	// Check if the key exists in the schema
	exists, err := r.KeyExistsInSchema(ctx, configKey)
	if err != nil {
		return "", fmt.Errorf("failed to check if key exists in schema: %w", err)
	}
	if !exists {
		return "", &KeyNotFoundError{Key: configKey}
	}

	// Construct the key for the parameter
	key := GetConfKeyPath(r.App, r.Module, r.Version, r.Config, configKey)

	// Try to get the value from the cache if caching is enabled
	if r.cacheEnabled.Load() {
		value, found := r.Cache.Get(key)
		if found {
			return value, nil
		}
	}

	// If the value is not in the cache or caching is disabled, retrieve it from the storage
	valueStr, err := r.Storage.Get(ctx, key)
	if err != nil {
		return "", &KeyNotFoundError{Key: key}
	}

	// Store the value in the cache if caching is enabled
	if r.cacheEnabled.Load() {
		r.Cache.Set(key, valueStr)
	}

	return valueStr, nil
}

func (r *Rigel) GetInt(ctx context.Context, configKey string) (int, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return 0, err
	}
	intValue, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to int: %w", err)
	}
	return intValue, nil
}

func (r *Rigel) GetFloat(ctx context.Context, configKey string) (float64, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return 0, err
	}
	floatValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to float: %w", err)
	}
	return floatValue, nil
}

func (r *Rigel) GetBool(ctx context.Context, configKey string) (bool, error) {
	valueStr, err := r.Get(ctx, configKey)
	if err != nil {
		return false, err
	}
	boolValue, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert value to bool: %w", err)
	}
	return boolValue, nil
}

func (r *Rigel) GetString(ctx context.Context, configKey string) (string, error) {
	valueStr, err := r.Get(ctx, configKey)
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
	case "float":
		floatValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to float: %w", err)
		}
		return floatValue, nil
	case "json":
		if !json.Valid([]byte(valueStr)) {
			return nil, fmt.Errorf("failed to validate json: invalid JSON")
		}
		return valueStr, nil
	default: // "string"
		return valueStr, nil
	}
}

// WatchConfig starts watching for changes to any key in the specified configuration namespace in the storage.
// When a change is detected, it updates the corresponding key-value pair in the cache if caching is enabled.
// The method takes the schemaName, schemaVersion, and configName
// to construct the base key for the configuration namespace.
// Note: When caching is disabled, events are still received but not cached.
// Security: Only keys that already exist in the cache are updated. New keys are not added to prevent
// cache pollution and ensure the cache only contains keys the application actually uses.
func (r *Rigel) WatchConfig(ctx context.Context) error {
	// Construct the base key for the configuration
	baseKey := GetConfPath(r.App, r.Module, r.Version, r.Config)

	events := make(chan types.Event)
	if err := r.Storage.Watch(ctx, baseKey, events); err != nil {
		return err
	}

	go func() {
		for event := range events {
			// Only update the cache if caching is enabled
			if r.cacheEnabled.Load() {
				// Only update keys in the cache that have changed
				if _, found := r.Cache.Get(event.Key); found {
					r.Cache.Set(event.Key, event.Value)
				}
			}
		}
	}()

	return nil
}
