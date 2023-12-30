package main

import (
	"log"

	"github.com/remiges-tech/rigel/cmd/rigelctl/rigelctl"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
	"github.com/spf13/cobra"
)

func main() {
	// Create a new EtcdStorage instance
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		log.Fatalf("Failed to create EtcdStorage: %v", err)
	}
	// Create a new Rigel instance
	rigelClient := rigel.New(etcdStorage, "testapp", "testmodule", 1, "testconfig")

	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "rigelctl",
		Short: "CLI for managing Rigel schemas and configs",
	}

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
