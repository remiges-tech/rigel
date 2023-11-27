// Package types defines the core data types used in Rigel.
package types

import (
	"context"
)

// Schema represents the structure of a schema. Currently, the only supported type is JSON.
//
// Example:
//
//	{
//	  "name": "webServer",
//	  "version": 1,
//	  "fields": [
//	    {"name": "host", "type": "string"},
//	    {"name": "port", "type": "int"},
//	    {"name": "logLevel", "type": "string"},
//	    {"name": "maxConnections", "type": "int"},
//	    {"name": "enableHttps", "type": "bool"},
//	  ],
//	  "description": "Configuration for a web server application"
//	}
type Schema struct {
	Name        string  // Name is the identifier of the schema
	Version     int     // Version is the version number of the schema
	Fields      []Field // Fields is a list of fields that the schema contains
	Description string  // Description provides more information about the schema
}

// Field represents a single field in a schema. Currently, the only supported types are string, int, and bool.
//
// Example:
//
//	{
//	  "name": "maxConnections",
//	  "type": "int"
//	}
type Field struct {
	Name string `json:"name"` // Name represents the name of the field (config parameter).
	Type string `json:"type"` // Type represents the type of the field. Currently, the supported types are "string", "int", and "bool".
}

// Storage is an interface that abstracts the operations for getting and putting data in
// Rigel's underlying storage
type Storage interface {
	// Get retrieves a value associated with the given key.
	// If the key does not exist, it returns an empty string and no error.
	// If an error occurs during the operation, it is returned.
	Get(ctx context.Context, key string) (string, error)

	// Put stores a value with the specified key.
	// If the key already exists, its value is updated; if it does not, a new key-value pair is created.
	// If an error occurs during the operation, it is returned.
	Put(ctx context.Context, key string, value string) error
}
