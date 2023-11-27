package rigelctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ssd532/rigel"
	"github.com/ssd532/rigel/types"
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

	// Call AddSchema
	err = client.AddSchema(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to add schema: %v", err)
	}

	return nil
}
