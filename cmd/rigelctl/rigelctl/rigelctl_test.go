package rigelctl

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/mocks"
	"github.com/spf13/cobra"
)

func TestAddSchemaFromFile(t *testing.T) {
	// Create a mock Rigel client
	mockRigelClient := &rigel.Rigel{
		Storage: &mocks.MockStorage{
			PutFunc: func(ctx context.Context, key string, value string) error {
				return nil
			},
		},
	}

	// Create a temporary file and write a sample schema to it
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	schema := `{
        "name": "webServer",
        "version": 1,
        "fields": [
            {"name": "host", "type": "string"},
            {"name": "port", "type": "int"},
            {"name": "enableHttps", "type": "bool"}
        ],
        "description": "Configuration for a web server application"
    }`
	if _, err := tmpfile.Write([]byte(schema)); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Run the addSchemaCmd with the temporary file as an argument
	cmd := &cobra.Command{}
	args := []string{tmpfile.Name()}
	err = AddSchemaCommand(mockRigelClient, cmd, args)
	if err != nil {
		t.Errorf("AddSchemaCommand with file argument failed: %v", err)
	}
}
