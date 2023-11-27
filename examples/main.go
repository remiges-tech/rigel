package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
)

type Config struct {
	DatabaseURL string `json:"database_url"`
	MaxRetries  int    `json:"max_retries"`
	EnableSSL   bool   `json:"enable_ssl"`
}

func main() {
	// Create a new EtcdStorage instance
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		log.Fatalf("Failed to create EtcdStorage: %v", err)
	}

	// Create a new Rigel instance
	rigelClient := rigel.New(etcdStorage)

	// Define a config struct
	var config Config

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Load the config
	err = rigelClient.LoadConfig(ctx, "appConfig", 1, "appConfig", &config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print the loaded config
	fmt.Printf("DatabaseURL: %s\n", config.DatabaseURL)
	fmt.Printf("MaxRetries: %d\n", config.MaxRetries)
	fmt.Printf("EnableSSL: %t\n", config.EnableSSL)
}
