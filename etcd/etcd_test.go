package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/remiges-tech/rigel/types"
	"go.etcd.io/etcd/tests/v3/integration"
)

func TestGetNonExistentKey(t *testing.T) {
	// Setup the test environment
	integration.BeforeTestExternal(t)

	// Create an embedded etcd server for testing
	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	// Create an EtcdStorage instance
	etcdStorage := &EtcdStorage{
		Client: clus.RandClient(),
	}

	// Try to get a non-existent key
	value, err := etcdStorage.Get(context.Background(), "non-existent-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned value is an empty string
	if value != "" {
		t.Errorf("Expected an empty string, got '%s'", value)
	}
}

func TestEtcdStorage_Watch(t *testing.T) {
	// Setup the test environment
	integration.BeforeTestExternal(t)

	// Create an embedded etcd server for testing
	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	// Create an EtcdStorage instance
	etcdStorage := &EtcdStorage{
		Client: clus.RandClient(),
	}

	// Create a channel for events
	events := make(chan types.Event)

	// Create a channel for errors
	errs := make(chan error)

	// Start watching a key
	go func() {
		err := etcdStorage.Watch(context.Background(), "test-key", events)
		if err != nil {
			errs <- err
		}
	}()

	// Check for any errors
	select {
	case err := <-errs:
		t.Fatalf("Expected no error, got %v", err)
	default:
	}

	// Put a value to the key
	err := etcdStorage.Put(context.Background(), "test-key", "test-value")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the event is received
	select {
	case event := <-events:
		if event.Key != "test-key" || event.Value != "test-value" {
			t.Errorf("Expected event with key 'test-key' and value 'test-value', got key '%s' and value '%s'", event.Key, event.Value)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Expected to receive an event, but didn't")
	}
}

func TestGetWithPrefix(t *testing.T) {
	// Setup the test environment
	integration.BeforeTestExternal(t)

	// Create an embedded etcd server for testing
	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	// Create an EtcdStorage instance
	etcdStorage := &EtcdStorage{
		Client: clus.RandClient(),
	}

	// Define a common prefix for Rigel-specific paths for version 1
	prefixV1 := "/remiges/rigel/testApp/testModule/1/"
	// Define a common prefix for Rigel-specific paths for version 2
	prefixV2 := "/remiges/rigel/testApp/testModule/2/"

	// Put two key-value pairs with the common prefix for version 1
	err := etcdStorage.Put(context.Background(), prefixV1+"key1", "value1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = etcdStorage.Put(context.Background(), prefixV1+"key2", "value2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Put two key-value pairs with the common prefix for version 2
	err = etcdStorage.Put(context.Background(), prefixV2+"key1", "value1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = etcdStorage.Put(context.Background(), prefixV2+"key2", "value2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Call GetWithPrefix with the common prefix for version 1
	keyVal, err := etcdStorage.GetWithPrefix(context.Background(), prefixV1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned map contains the correct keys and values for version 1
	if keyVal[prefixV1+"key1"] != "value1" || keyVal[prefixV1+"key2"] != "value2" {
		t.Errorf("Expected map with keys '%skey1' and '%skey2' and corresponding values, got %v", prefixV1, prefixV1, keyVal)
	}

	// Check if the returned map does not contain keys for version 2
	if _, ok := keyVal[prefixV2+"key1"]; ok {
		t.Errorf("Did not expect to find key '%skey1'", prefixV2)
	}
	if _, ok := keyVal[prefixV2+"key2"]; ok {
		t.Errorf("Did not expect to find key '%skey2'", prefixV2)
	}
}
