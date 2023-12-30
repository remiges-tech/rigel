package main

import (
	"context"
	"fmt"
	"time"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
)

func main() {
	// Connect to etcd
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		fmt.Printf("Failed to connect to etcd: %v\n", err)
		return
	}

	// Create a new Rigel instance
	r := rigel.New(etcdStorage, "testapp", "testmodule", 1, "testconfig")

	// Start watching for changes
	err = r.WatchConfig(context.Background())
	if err != nil {
		fmt.Printf("Failed to watch config: %v\n", err)
		return
	}

	// Keep the program running and print the cache contents whenever a change is detected
	var prevValue string
	for {
		fmt.Printf("cache: %v", r.Cache)
		time.Sleep(time.Second)
		value, err := r.Get(context.Background(), "maxAge")
		if err != nil {
			fmt.Printf("Failed to get value: %v\n", err)
			continue
		}
		fmt.Printf("Current value: %s\n", value)
		if value != prevValue {
			fmt.Printf("Config has been updated: %s\n", value)
			prevValue = value
		}
	}
}
