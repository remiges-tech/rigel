package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unsafe"

	"github.com/remiges-tech/rigel/cmd/rigelctl/rigelctl"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
	"github.com/spf13/cobra"
)

func main() {
	var etcdEndpoint, app, module string

	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "rigelctl",
		Short: "CLI for managing Rigel schemas and configs",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Split the etcdEndpoint string into a slice of strings
			etcdEndpoints := strings.Split(etcdEndpoint, ",")

			// Create a new EtcdStorage instance
			etcdStorage, err := etcd.NewEtcdStorage(etcdEndpoints)
			if err != nil {
				return fmt.Errorf("Failed to create EtcdStorage: %v", err)
			}

			// Create a new Rigel instance with the provided Storage interface
			rigelClient := rigel.NewWithStorage(etcdStorage)

			// Set the App and Module fields using the WithApp and WithModule methods
			rigelClient = rigelClient.WithApp(app).WithModule(module)

			// Store the Rigel client pointer in the command's annotations for later retrieval
			cmd.Annotations = make(map[string]string)
			cmd.Annotations["rigelClient"] = fmt.Sprintf("%p", rigelClient)
			return nil
		},
	}
	rootCmd.PersistentFlags().StringVarP(&etcdEndpoint, "etcd-endpoint", "e", "localhost:2379", "etcd endpoint")
	rootCmd.PersistentFlags().StringVarP(&app, "app", "a", "testapp", "app name")
	rootCmd.PersistentFlags().StringVarP(&module, "module", "m", "testmodule", "module name")

	// Create the 'schema' command
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage Rigel schemas",
	}

	// Create the 'add' command under 'schema'
	addSchemaCmd := &cobra.Command{
		Use:   "add [schema_file]",
		Short: "Add a new schema from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rigelClientPtr, _ := strconv.ParseUint(cmd.Annotations["rigelClient"], 0, 64)
			rigelClient := (*rigel.Rigel)(unsafe.Pointer(uintptr(rigelClientPtr)))

			// Check if the rigelClient is nil
			if rigelClient == nil {
				return fmt.Errorf("Failed to initialize Rigel client")
			}

			return rigelctl.AddSchemaCommand(rigelClient, cmd, args)
		},
	}
	// Add the 'addSchema' command to the 'schema' command
	schemaCmd.AddCommand(addSchemaCmd)

	// Add the 'schema' command to the root command
	rootCmd.AddCommand(schemaCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute root command: %v", err)
	}
}
