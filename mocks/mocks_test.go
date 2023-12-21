package mocks

import (
	"context"
	"fmt"
	"testing"

	"github.com/remiges-tech/rigel/types"
)

func ExampleMockStorage_Get() {
	// Create a new MockStorage instance
	mockStorage := &MockStorage{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == "testKey" {
				return "testValue", nil
			}
			return "", fmt.Errorf("unexpected key: %s", key)
		},
	}

	// Call the Get method with "testKey"
	value, err := mockStorage.Get(context.Background(), "testKey")
	if err != nil {
		fmt.Printf("Expected no error, got %v\n", err)
		return
	}

	// Check the returned value
	if value != "testValue" {
		fmt.Printf("Expected value to be 'testValue', got '%s'\n", value)
		return
	}

	// Print the value
	fmt.Println(value)
	// Output: testValue
}

func TestMockStorage_Watch(t *testing.T) {
	// Define the expected event
	expectedEvent := types.Event{
		Key:   "testKey",
		Value: "testValue",
	}

	// Create a new MockStorage instance
	mockStorage := &MockStorage{
		WatchFunc: func(ctx context.Context, key string, events chan<- types.Event) error {
			// Send the expected event to the channel
			events <- expectedEvent
			return nil
		},
	}

	// Create a channel to receive events
	events := make(chan types.Event, 1)

	// Call the Watch method with "testKey"
	err := mockStorage.Watch(context.Background(), "testKey", events)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Receive the event from the channel
	event := <-events

	// Check the received event
	if event != expectedEvent {
		t.Errorf("Expected event to be '%v', got '%v'", expectedEvent, event)
	}
}
