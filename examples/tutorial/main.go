package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create etcd storage
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		log.Fatalf("Failed to create etcd storage: %v", err)
	}

	// Initialize Rigel client with all parameters
	rigelClient := rigel.New(etcdStorage, "alya", "usersvc", 1, "dev")

	fmt.Println("Reading configuration values:")

	// Database configuration
	dbHost, err := rigelClient.Get(ctx, "database.host")
	if err != nil {
		log.Printf("Failed to get database host: %v", err)
	}
	fmt.Printf("Database Host: %s\n", dbHost)

	dbPort, err := rigelClient.GetInt(ctx, "database.port")
	if err != nil {
		log.Printf("Failed to get database port: %v", err)
	}
	fmt.Printf("Database Port: %v\n", dbPort)

	dbUser, err := rigelClient.Get(ctx, "database.user")
	if err != nil {
		log.Printf("Failed to get database user: %v", err)
	}
	fmt.Printf("Database User: %s\n", dbUser)

	// Server configuration
	serverPort, err := rigelClient.Get(ctx, "server.port")
	if err != nil {
		log.Printf("Failed to get server port: %v", err)
	}
	fmt.Printf("Server Port: %s\n", serverPort)
}
