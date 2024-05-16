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
    "fields": [
        {
            "name": "host",
            "type": "string",
            "description": "The hostname or IP address of the web server."
        },
        {
            "name": "port",
            "type": "int",
            "description": "The port number on which the web server listens for incoming requests.",
            "constraints": {
                "min": 1,
                "max": 65535
            }
        },
        {
            "name": "enableHttps",
            "type": "bool",
            "description": "Indicates whether HTTPS should be enabled for secure communication."
        }
    ],
    "description": "Configuration schema for a web server application."
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

func TestValidateSchema(t *testing.T) {
	validSchema := []byte(`{
		"fields": [
			{
				"name": "transactionTimeout",
				"type": "int",
				"description": "Defines the maximum duration (in seconds) a transaction should take before timing out.",
				"constraints": {
					"min": 1,
					"max": 60
				}
			}
		],
		"description": "Configuration schema of the PaymentGateway module in FinanceApp."
	}`)

	invalidSchemaNoDescription := []byte(`{
		"fields": [
			{
				"name": "transactionTimeout",
				"type": "int"
			}
		]
	}`)

	invalidSchemaConstraint := []byte(`{
		"fields": [
			{
				"name": "transactionTimeout",
				"type": "int",
				"description": "Defines the maximum duration (in seconds) a transaction should take before timing out.",
				"constraints": {
					"mean": 30
				}
			}
		],
		"description": "Configuration schema of the PaymentGateway module in FinanceApp."
	}`)

	tests := []struct {
		name        string
		schemaBytes []byte
		wantErr     bool
	}{
		{"Valid Schema", validSchema, false},
		{"Invalid Schema: No description", invalidSchemaNoDescription, true},
		{"Invalid Schema: Invalid constraint", invalidSchemaConstraint, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchema(tt.schemaBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
