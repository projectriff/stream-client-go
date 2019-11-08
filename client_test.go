package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	client "github.com/projectriff/stream-client-go"
)

// This is an integration test meant to be run against a kafka custer. Please refer to the CI scripts for
// setup details
func TestNewStreamClient(t *testing.T) {
	Publish("FOO", "text/plain", "test_topic", t)
	Subscribe("FOO", "test_topic", t)
}

func Publish(value, contentType, topic string, t *testing.T) {
	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}
	headers := make(map[string]string)
	reader := strings.NewReader(value)
	publishResult, err := c.Publish(context.Background(), reader, nil, contentType, headers)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Published: %+v", publishResult)
}

func Subscribe(expectedValue, topic string, t *testing.T) {
	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}

	var errHandler client.SubscribeErrHandler
	errHandler = func(cancel context.CancelFunc, err error) {
		fmt.Printf("cancelling subsctiber due to: %v", err)
		cancel()
	}

	result := make(chan string, 2)

	var subscriber client.Subscriber
	subscriber = func(ctx context.Context, payload []byte, contentType string) error {
		select {
		case <- ctx.Done():
			return ctx.Err()
		default:
		}

		result <- string(payload)
		return nil
	}

	_, err = c.Subscribe(context.Background(), "g8", subscriber, errHandler)
	if err != nil {
		t.Error(err)
	}
	v1 := <- result
	if v1 != expectedValue {
		t.Errorf("expected value: %s, but was: %s", expectedValue, v1)
	}
}
