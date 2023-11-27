package etcd

import (
	"context"
	"testing"

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
