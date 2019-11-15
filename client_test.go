package client_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"time"

	client "github.com/projectriff/stream-client-go"
)

// This is an integration test meant to be run against a kafka custer. Please refer to the CI scripts for
// setup details
func TestSimplePublishSubscribe(t *testing.T) {
	now := time.Now()
	topic := fmt.Sprintf("test_%s%d%d%d", t.Name(), now.Hour(), now.Minute(), now.Second())

	c := setupStreamingClient(topic, t)

	payload := "FOO"
	headers := map[string]string{"H1":"V1", "H2":"V2"}
	publish(c, payload, "text/plain", topic, headers, t)
	subscribe(c, payload, topic, 0, headers, t)
}

func setupStreamingClient(topic string, t *testing.T) *client.StreamClient {
	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}
	return c
}

func publish(c *client.StreamClient, value, contentType, topic string, headers map[string]string, t *testing.T) {
	reader := strings.NewReader(value)
	publishResult, err := c.Publish(context.Background(), reader, nil, contentType, headers)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Published: %+v\n", publishResult)
}

func subscribe(c *client.StreamClient, expectedValue, topic string, offset uint64, headers map[string]string, t *testing.T) {

	var errHandler client.EventErrHandler
	errHandler = func(cancel context.CancelFunc, err error) {
		fmt.Printf("cancelling subsctiber due to: %v", err)
		cancel()
	}

	payloadChan := make(chan string)
	headersChan := make(chan map[string]string)

	var eventHandler client.EventHandler
	eventHandler = func(ctx context.Context, payload io.Reader, contentType string, headers map[string]string) error {
		bytes, err := ioutil.ReadAll(payload)
		if err != nil {
			return err
		}
		payloadChan <- string(bytes)
		headersChan <- headers
		return nil
	}

	_, err := c.Subscribe(context.Background(), "g8", offset, eventHandler, errHandler)
	if err != nil {
		t.Error(err)
	}
	v1 := <-payloadChan
	if v1 != expectedValue {
		t.Errorf("expected value: %s, but was: %s", expectedValue, v1)
	}
	h := <- headersChan
	if !reflect.DeepEqual(headers, h) {
		t.Errorf("headers not equal. expected %s, but was: %s", headers, h)
	}
}

func TestSubscribeBeforePublish(t *testing.T) {
	now := time.Now()
	topic := fmt.Sprintf("test_%s%d%d%d", t.Name(), now.Hour(), now.Minute(), now.Second())

	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}

	testVal := "BAR"
	result := make(chan string)

	var eventHandler client.EventHandler
	eventHandler = func(ctx context.Context, payload io.Reader, contentType string, headers map[string]string) error {
		bytes, err := ioutil.ReadAll(payload)
		if err != nil {
			return err
		}
		result <- string(bytes)
		return nil
	}
	var eventErrHandler client.EventErrHandler
	eventErrHandler = func(cancel context.CancelFunc, err error) {
		t.Error("Did not expect an error")
	}
	_, err = c.Subscribe(context.Background(), t.Name(), 0, eventHandler, eventErrHandler)
	if err != nil {
		t.Error(err)
	}
	publish(c, testVal, "text/plain", topic, nil, t)
	v1 := <- result
	if v1 != testVal {
		t.Errorf("expected value: %s, but was: %s", testVal, v1)
	}
}

func TestSubscribeCancel(t *testing.T) {
	now := time.Now()
	topic := fmt.Sprintf("test_%s%d%d%d", t.Name(), now.Hour(), now.Minute(), now.Second())

	c, err := client.NewStreamClient("localhost:6565", topic, "text/plain")
	if err != nil {
		t.Error(err)
	}

	expectedError := "expected_error"
	result := make(chan string)

	var eventHandler client.EventHandler
	eventHandler = func(ctx context.Context, payload io.Reader, contentType string, headers map[string]string) error {
		bytes, err := ioutil.ReadAll(payload)
		if err != nil {
			return err
		}
		result <- string(bytes)
		return nil
	}
	var eventErrHandler client.EventErrHandler
	eventErrHandler = func(cancel context.CancelFunc, err error) {
		result <- expectedError
	}
	cancel, err := c.Subscribe(context.Background(), t.Name(), 0, eventHandler, eventErrHandler)
	if err != nil {
		t.Error(err)
	}
	cancel()
	v1 := <- result
	if v1 != expectedError {
		t.Errorf("expected value: %s, but was: %s", expectedError, v1)
	}
}
