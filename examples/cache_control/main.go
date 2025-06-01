package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
	"github.com/remiges-tech/rigel/types"
)

const (
	testApp     = "demoapp"
	testModule  = "cachecontrol"
	testVersion = 1
	testConfig  = "production"
)

func main() {
	// Connect to etcd
	etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}

	// Create a new Rigel instance
	client := rigel.New(etcdStorage, testApp, testModule, testVersion, testConfig)
	ctx := context.Background()

	// Setup schema for demo
	err = setupDemoSchema(ctx, client)
	if err != nil {
		log.Fatalf("Failed to setup schema: %v", err)
	}

	// Initialize some demo configuration values
	err = initializeDemoConfig(ctx, client)
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	fmt.Println("=== Rigel Cache Control Demo ===")

	// Demo 1: Basic cache control
	demonstrateBasicCacheControl(ctx, client)

	// Demo 2: Performance comparison
	demonstratePerformanceComparison(ctx, client)

	// Demo 3: Debugging scenario
	demonstrateDebuggingScenario(ctx, client)

	// Demo 4: Health check pattern
	demonstrateHealthCheck(ctx, client)

	// Demo 5: Critical operation pattern
	demonstrateCriticalOperation(ctx, client)

	// Demo 6: Schema cache invalidation
	demonstrateSchemaCacheInvalidation(ctx, client)

	fmt.Println("\n=== Demo Complete ===")
}

func setupDemoSchema(ctx context.Context, client *rigel.Rigel) error {
	schema := types.Schema{
		Version:     testVersion,
		Description: "Demo schema for cache control examples",
		Fields: []types.Field{
			{Name: "api_endpoint", Type: "string", Description: "API endpoint URL"},
			{Name: "timeout_seconds", Type: "int", Description: "Request timeout in seconds"},
			{Name: "debug_mode", Type: "bool", Description: "Enable debug logging"},
			{Name: "rate_limit", Type: "float", Description: "Rate limit per second"},
			{Name: "health_check_key", Type: "string", Description: "Health check value"},
		},
	}

	return client.AddSchema(ctx, schema)
}

func initializeDemoConfig(ctx context.Context, client *rigel.Rigel) error {
	configs := map[string]string{
		"api_endpoint":     "https://api.example.com",
		"timeout_seconds":  "30",
		"debug_mode":       "false",
		"rate_limit":       "100.5",
		"health_check_key": "healthy",
	}

	for key, value := range configs {
		if err := client.Set(ctx, key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

func demonstrateBasicCacheControl(ctx context.Context, client *rigel.Rigel) {
	fmt.Println("1. Basic Cache Control & Timing")
	fmt.Println("-------------------------------")

	// Check default state
	fmt.Printf("Default caching state: %v\n", client.IsCachingEnabled())

	// Clear cache and schema cache
	client.DisableCaching().EnableCaching()
	// Clear value cache by getting a non-existent key to avoid pre-population
	client.DisableCaching()
	client.InvalidateSchemaCache()
	client.EnableCaching()

	// First call - populates cache
	start := time.Now()
	endpoint, _ := client.Get(ctx, "api_endpoint")
	firstCallDuration := time.Since(start)
	fmt.Printf("First call (populates cache): %s [%v]\n", endpoint, firstCallDuration)

	// Second call - from cache
	start = time.Now()
	endpoint, _ = client.Get(ctx, "api_endpoint")
	cachedDuration := time.Since(start)
	fmt.Printf("Second call (from cache): %s [%v]\n", endpoint, cachedDuration)

	// Calculate speedup
	if cachedDuration > 0 {
		speedup := float64(firstCallDuration) / float64(cachedDuration)
		fmt.Printf("Cache speedup for single call: %.1fx faster\n", speedup)
	}

	// Disable caching and measure
	client.DisableCaching()
	fmt.Printf("\nAfter DisableCaching(): %v\n", client.IsCachingEnabled())

	// Multiple uncached calls to show consistency
	fmt.Println("Multiple calls without cache:")
	for i := 1; i <= 3; i++ {
		start = time.Now()
		endpoint, _ = client.Get(ctx, "api_endpoint")
		duration := time.Since(start)
		fmt.Printf("  Call %d: %s [%v]\n", i, endpoint, duration)
	}

	// Re-enable caching
	client.EnableCaching()

	// First call after re-enabling (populates cache again)
	start = time.Now()
	endpoint, _ = client.Get(ctx, "api_endpoint")
	reenabledDuration := time.Since(start)
	fmt.Printf("After re-enabling (repopulates cache): %s [%v]\n", endpoint, reenabledDuration)

	// Second call to show it's now cached
	start = time.Now()
	endpoint, _ = client.Get(ctx, "api_endpoint")
	cachedAgainDuration := time.Since(start)
	fmt.Printf("Subsequent call (from cache again): %s [%v]\n", endpoint, cachedAgainDuration)

	// Demonstrate method chaining with timing
	fmt.Println("\nMethod chaining demonstration:")
	start = time.Now()
	timeout, _ := client.DisableCaching().GetInt(ctx, "timeout_seconds")
	chainDuration := time.Since(start)
	client.EnableCaching()
	fmt.Printf("Timeout via chained DisableCaching: %d [%v]\n", timeout, chainDuration)

	fmt.Println()
}

func demonstratePerformanceComparison(ctx context.Context, client *rigel.Rigel) {
	fmt.Println("2. Performance Comparison")
	fmt.Println("-------------------------")

	const iterations = 100

	// Measure cached performance
	client.EnableCaching()
	// Pre-populate cache
	client.Get(ctx, "api_endpoint")

	start := time.Now()
	for i := 0; i < iterations; i++ {
		client.Get(ctx, "api_endpoint")
	}
	cachedDuration := time.Since(start)

	// Measure non-cached performance
	client.DisableCaching()
	start = time.Now()
	for i := 0; i < iterations; i++ {
		client.Get(ctx, "api_endpoint")
	}
	noCacheDuration := time.Since(start)

	client.EnableCaching() // Restore default

	fmt.Printf("Cached (%d calls): %v\n", iterations, cachedDuration)
	fmt.Printf("No cache (%d calls): %v\n", iterations, noCacheDuration)
	if noCacheDuration > 0 {
		speedup := float64(noCacheDuration) / float64(cachedDuration)
		fmt.Printf("Cache speedup: %.1fx faster\n", speedup)
	}

	fmt.Println()
}

func demonstrateDebuggingScenario(ctx context.Context, client *rigel.Rigel) {
	fmt.Println("3. Debugging Scenario - Cache Staleness Detection")
	fmt.Println("-------------------------------------------------")

	// Setup: Enable caching and get initial value
	client.EnableCaching()
	fmt.Println("Initial setup:")
	initialValue, _ := client.Get(ctx, "debug_mode")
	fmt.Printf("  Initial value (cached): %s\n", initialValue)

	// Simulate external change: Another process updates the value
	fmt.Println("\nSimulating external configuration change...")
	// Directly update etcd bypassing cache
	newValue := "true"
	if initialValue == "true" {
		newValue = "false"
	}
	err := client.Set(ctx, "debug_mode", newValue)
	if err != nil {
		fmt.Printf("  Error updating value: %v\n", err)
	} else {
		fmt.Printf("  Value changed in etcd to: %s\n", newValue)
	}

	// Compare cached vs fresh values
	fmt.Println("\nInvestigating cache state:")

	// Get potentially stale cached value
	start := time.Now()
	cachedValue, _ := client.Get(ctx, "debug_mode")
	cachedTime := time.Since(start)
	fmt.Printf("  Cached value: %s [%v]\n", cachedValue, cachedTime)

	// Get fresh value for comparison
	client.DisableCaching()
	start = time.Now()
	freshValue, _ := client.Get(ctx, "debug_mode")
	freshTime := time.Since(start)
	fmt.Printf("  Fresh value from etcd: %s [%v]\n", freshValue, freshTime)

	// Analyze results
	if cachedValue == freshValue {
		fmt.Println("\nResult: Cache is up to date")
	} else {
		fmt.Println("\nResult: Cache is STALE!")
		fmt.Println("Action: Consider using WatchConfig() for real-time updates")

		// Show how to fix
		fmt.Println("\nFixing stale cache:")
		client.EnableCaching()
		// Force cache refresh by disabling and re-enabling
		client.DisableCaching().EnableCaching()
		fixedValue, _ := client.Get(ctx, "debug_mode")
		fmt.Printf("  After cache refresh: %s (now matches etcd)\n", fixedValue)
	}

	client.EnableCaching() // Restore
	fmt.Println()
}

func demonstrateHealthCheck(ctx context.Context, client *rigel.Rigel) {
	fmt.Println("4. Health Check Pattern")
	fmt.Println("-----------------------")

	// Health check function that ensures etcd connectivity
	healthCheck := func() error {
		// Temporarily disable cache to test actual etcd connection
		originalState := client.IsCachingEnabled()
		client.DisableCaching()
		defer func() {
			if originalState {
				client.EnableCaching()
			}
		}()

		// This will fail if etcd is unreachable
		_, err := client.Get(ctx, "health_check_key")
		return err
	}

	fmt.Println("Running health check (bypasses cache)...")
	if err := healthCheck(); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
	} else {
		fmt.Println("Health check passed - etcd is reachable")
	}

	fmt.Println()
}

func demonstrateCriticalOperation(ctx context.Context, client *rigel.Rigel) {
	fmt.Println("5. Critical Operation Pattern")
	fmt.Println("-----------------------------")

	// Simulate a critical operation that needs guaranteed fresh values
	processPayment := func(amount float64) error {
		fmt.Printf("Processing payment of $%.2f\n", amount)

		// Disable cache for critical rate limit check
		client.DisableCaching()
		defer client.EnableCaching()

		// Get the latest rate limit (guaranteed fresh)
		rateLimit, err := client.GetFloat(ctx, "rate_limit")
		if err != nil {
			return fmt.Errorf("failed to get rate limit: %w", err)
		}

		fmt.Printf("Current rate limit: %.1f/sec (fresh from etcd)\n", rateLimit)

		if amount > rateLimit*10 { // Simple validation rule
			return fmt.Errorf("amount exceeds rate limit threshold")
		}

		fmt.Println("Payment processed successfully")
		return nil
	}

	// Process a payment
	if err := processPayment(50.0); err != nil {
		fmt.Printf("Payment failed: %v\n", err)
	}

	fmt.Println()
}

func demonstrateSchemaCacheInvalidation(ctx context.Context, client *rigel.Rigel) {
	fmt.Println("6. Schema Cache Management")
	fmt.Println("--------------------------")

	// Enable caching to use schema cache
	client.EnableCaching()

	// First access - loads schema into cache
	fmt.Println("Testing schema cache performance:")
	start := time.Now()
	_, err := client.Get(ctx, "api_endpoint")
	firstSchemaDuration := time.Since(start)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  First access (loads schema + value): %v\n", firstSchemaDuration)

	// Second access - uses cached schema
	start = time.Now()
	_, _ = client.Get(ctx, "timeout_seconds")
	cachedSchemaDuration := time.Since(start)
	fmt.Printf("  Second access (cached schema): %v\n", cachedSchemaDuration)

	// Show the benefit
	if cachedSchemaDuration > 0 {
		schemaSpeedup := float64(firstSchemaDuration) / float64(cachedSchemaDuration)
		fmt.Printf("  Schema cache benefit: %.1fx faster\n", schemaSpeedup)
	}

	// Demonstrate manual schema invalidation
	fmt.Println("\nManual schema cache invalidation:")
	client.InvalidateSchemaCache()
	fmt.Println("  Schema cache cleared")

	start = time.Now()
	_, _ = client.Get(ctx, "debug_mode")
	reloadDuration := time.Since(start)
	fmt.Printf("  Next access (reloads schema): %v\n", reloadDuration)

	// Show that DisableCaching also clears schema cache
	fmt.Println("\nDisableCaching() effect on schema cache:")
	client.EnableCaching()
	_, _ = client.Get(ctx, "rate_limit") // Populate cache

	start = time.Now()
	_, _ = client.Get(ctx, "rate_limit")
	beforeDisable := time.Since(start)
	fmt.Printf("  Before DisableCaching: %v\n", beforeDisable)

	client.DisableCaching().EnableCaching()
	start = time.Now()
	_, _ = client.Get(ctx, "rate_limit")
	afterDisable := time.Since(start)
	fmt.Printf("  After DisableCaching->EnableCaching: %v (schema reloaded)\n", afterDisable)

	fmt.Println()
}
