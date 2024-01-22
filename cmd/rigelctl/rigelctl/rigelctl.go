package rigelctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/types"
	"github.com/spf13/cobra"
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

	// Parse the schema from the file
	var schema types.Schema
	err = json.Unmarshal(fileBytes, &schema)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// Call AddSchema
	err = client.AddSchema(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to add schema: %v", err)
	}

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
