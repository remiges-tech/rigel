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
	var etcdEndpoint, app, module, config string
	var version int

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
	rootCmd.PersistentFlags().StringVarP(&app, "app", "a", "", "app name")
	rootCmd.PersistentFlags().StringVarP(&module, "module", "m", "", "module name")
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "config name")
	rootCmd.PersistentFlags().IntVarP(&version, "version", "v", 0, "version number")

	//
	// schema command
	//

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
			// Check if the required flags are provided
			if app == "" || module == "" {
				return fmt.Errorf("the 'app' and 'module' flags must be provided")
			}

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

	//
	// config command
	//

	// Create the 'config' command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Rigel configs",
	}

	// Create the 'set' command under 'config'
	setConfigCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a config key and its value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the required flags are provided
			if app == "" || module == "" || version == 0 || config == "" {
				return fmt.Errorf("the 'app', 'module', 'version', and 'config' flags must be provided")
			}

			// Retrieve the Rigel client from the command's annotations
			rigelClientPtr, _ := strconv.ParseUint(cmd.Annotations["rigelClient"], 0, 64)
			rigelClient := (*rigel.Rigel)(unsafe.Pointer(uintptr(rigelClientPtr)))

			// Check if the rigelClient is nil
			if rigelClient == nil {
				return fmt.Errorf("Failed to initialize Rigel client")
			}

			// Set the version and config name on the rigelClient
			rigelClient = rigelClient.WithVersion(version).WithConfig(config)

			// Call the SetConfigCommand function in the rigelctl package
			key := args[0]
			value := args[1]
			return rigelctl.SetConfigCommand(rigelClient, key, value)
		},
	}
	// Add the 'setConfig' command to the 'config' command
	configCmd.AddCommand(setConfigCmd)

	// Create the 'get' command under 'config'
	getConfigCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a config key value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the required flags are provided
			if app == "" || module == "" || version == 0 || config == "" {
				return fmt.Errorf("the 'app', 'module', 'version', and 'config' flags must be provided")
			}

			// Retrieve the Rigel client from the command's annotations
			rigelClientPtr, _ := strconv.ParseUint(cmd.Annotations["rigelClient"], 0, 64)
			rigelClient := (*rigel.Rigel)(unsafe.Pointer(uintptr(rigelClientPtr)))

			// Check if the rigelClient is nil
			if rigelClient == nil {
				return fmt.Errorf("Failed to initialize Rigel client")
			}

			// Set the version and config name on the rigelClient
			rigelClient = rigelClient.WithVersion(version).WithConfig(config)

			// Call the GetConfigCommand function in the rigelctl package
			key := args[0]
			return rigelctl.GetConfigCommand(rigelClient, key)
		},
	}

	// Add the 'getConfig' command to the 'config' command
	configCmd.AddCommand(getConfigCmd)

	// Add the 'config' command to the root command
	rootCmd.AddCommand(configCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute root command: %v", err)
	}
}
