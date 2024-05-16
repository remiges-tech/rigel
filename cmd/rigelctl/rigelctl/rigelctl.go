package rigelctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/types"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

func AddSchemaCommand(client *rigel.Rigel, cmd *cobra.Command, args []string) error {
	// Check if the file path argument is provided
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	filePath := args[0]

	// Read the file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Validate the schema
	err = ValidateSchema(fileBytes)
	if err != nil {
		return err
	}

	// Parse the schema from the file
	var schema types.Schema
	err = json.Unmarshal(fileBytes, &schema)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %v", err)
	}

	schema.Version = client.Version

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// Call AddSchema
	err = client.AddSchema(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to add schema: %v", err)
	}

	fmt.Println("Schema added successfully.")
	fmt.Printf("app: %s \nmodule: %s \nversion: %d\n", client.App, client.Module, client.Version)

	return nil
}

func SetConfigCommand(client *rigel.Rigel, key string, value string) error {
	// Set the config key and its value using the Set function
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := client.Set(ctx, key, value)
	if err != nil {
		return fmt.Errorf("Failed to set config: %v", err)
	}

	fmt.Printf("Config key '%s' set to '%s' successfully\n", key, value)
	return nil
}

func GetConfigCommand(client *rigel.Rigel, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	value, err := client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("Failed to get config: %v", err)
	}

	fmt.Printf("%s\n", value)
	return nil
}

func ValidateSchema(schemaBytes []byte) error {
	schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
	jsonSchemaLoader := gojsonschema.NewStringLoader(RigelSchemaJSON)

	result, err := gojsonschema.Validate(jsonSchemaLoader, schemaLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %v", err)
	}

	if !result.Valid() {
		var errMessages []string
		for _, err := range result.Errors() {
			errMessages = append(errMessages, err.String())
		}
		return fmt.Errorf("invalid schema:\n%s", strings.Join(errMessages, "\n"))
	}

	return nil
}
