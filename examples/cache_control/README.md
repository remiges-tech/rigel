# Cache Control Example

This example demonstrates how to use Rigel's cache control functionality to manage when configuration values are cached vs fetched fresh from etcd.

## Overview

Rigel provides runtime cache control through three methods:
- `EnableCaching()` - Enable caching (default behavior)
- `DisableCaching()` - Disable caching, all Gets fetch directly from etcd
- `IsCachingEnabled()` - Check current caching state

## Running the Example

### 1. Start etcd

```bash
docker-compose up -d
```

### 2. Run the Example

```bash
cd examples/cache_control
go run main.go
```

### 3. Clean up

```bash
docker-compose down

# To also remove the data volume
docker volume rm rigel-cache-control-etcd-data
```

## What the Example Demonstrates

### 1. Basic Cache Control
- Default caching state (enabled)
- Enabling/disabling cache
- Method chaining support

### 2. Performance Comparison
- Measures performance difference between cached and non-cached operations
- Shows cache speedup benefits

### 3. Debugging Scenario
- Compare cached vs fresh values
- Useful for investigating stale cache issues

### 4. Health Check Pattern
- Bypass cache to verify etcd connectivity
- Restore original cache state after check

### 5. Critical Operation Pattern
- Disable cache for operations requiring guaranteed fresh data
- Example: Payment processing with fresh rate limits

## Use Cases for Cache Control

### When to Disable Caching

1. **Debugging Production Issues**
   ```go
   client.DisableCaching()
   freshValue := client.Get("problematic_key")
   client.EnableCaching()
   ```

2. **Health Checks**
   ```go
   func healthCheck() error {
       client.DisableCaching()
       defer client.EnableCaching()
       _, err := client.Get("health_key")
       return err
   }
   ```

3. **Critical Operations**
   ```go
   func processPayment() {
       client.DisableCaching()
       defer client.EnableCaching()
       limit := client.GetFloat("payment_limit")
       // Process with guaranteed fresh limit
   }
   ```

4. **Configuration Auditing**
   ```go
   func auditConfig() {
       client.DisableCaching()
       defer client.EnableCaching()
       
       for _, key := range criticalKeys {
           etcdValue := client.Get(key)
           validateValue(key, etcdValue)
       }
   }
   ```

### When to Keep Caching Enabled

- Normal application operation (default)
- High-frequency config reads
- When performance is more important than freshness
- When using WatchConfig for real-time updates

## Performance Considerations

- **Cached operations**: ~0.5-10 microseconds (in-memory access)
- **Non-cached operations**: 
  - Local etcd: ~0.5-2 milliseconds
  - Remote etcd (same datacenter): ~1-5 milliseconds  
  - Remote etcd (cross-region): ~10-100 milliseconds
- **Typical speedup**: 
  - Local etcd: 50-500x
  - Remote etcd: 100-10,000x
  
Note: The speedup depends heavily on network latency between your application and etcd.

## Thread Safety

All cache control methods are thread-safe using atomic operations:
- No locks needed for concurrent access
- Safe to call from multiple goroutines
- High performance for read-heavy workloads