/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"github.com/projectriff/stream-client-go/pkg/liiklus"
	"github.com/projectriff/stream-client-go/pkg/serialization"
)

// StreamClient allows publishing to a riff stream, through a liiklus gateway and using the riff serialization format.
type StreamClient struct {
	// Gateway is the host:port of the liiklus gRPC endpoint.
	Gateway string
	// TopicName is the name of the liiklus topic backing the stream.
	TopicName string
	// acceptableContentType is the content type that the stream is able to persist. Incompatible content types will be rejected.
	acceptableContentType string
	// client is the gRPC client for the liiklus API.
	client liiklus.LiiklusServiceClient
	// conn is a reference to the underlying connection, kept for proper cleanup.
	conn *grpc.ClientConn
}

type PublishResult struct {
	Partition uint32
	Offset    uint64
}

// NewStreamClient creates a new StreamClient for a given stream.
func NewStreamClient(gateway string, topic string, acceptableContentType string) (*StreamClient, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	conn, err := grpc.DialContext(timeout, gateway, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	client := liiklus.NewLiiklusServiceClient(conn)
	return &StreamClient{
		Gateway:               gateway,
		TopicName:             topic,
		acceptableContentType: acceptableContentType,
		client:                client,
		conn:                  conn,
	}, nil
}

func (lc *StreamClient) Publish(ctx context.Context, payload io.Reader, key io.Reader, contentType string, headers map[string]string) (PublishResult, error) {
	m := serialization.Message{}
	if chopContentType(contentType) != chopContentType(lc.acceptableContentType) { // TODO support smarter compatibility (eg subtypes)
		return PublishResult{}, fmt.Errorf("contentType %q not compatible with expected contentType %q", contentType, lc.acceptableContentType)
	}
	m.ContentType = contentType
	if bytes, err := ioutil.ReadAll(payload); err != nil {
		return PublishResult{}, err
	} else {
		m.Payload = bytes
	}
	for k, v := range headers {
		m.Headers[k] = v
	}

	var err error
	var value []byte
	var kValue []byte
	if value, err = proto.Marshal(&m); err != nil {
		return PublishResult{}, err
	}
	if key != nil {
		if kValue, err = ioutil.ReadAll(key); err != nil {
			return PublishResult{}, err
		}
	}
	request := liiklus.PublishRequest{
		Topic: lc.TopicName,
		Value: value,
		Key:   kValue,
	}
	if publishReply, err := lc.client.Publish(ctx, &request); err != nil {
		return PublishResult{}, err
	} else {
		return PublishResult{Offset: publishReply.Offset, Partition: publishReply.Partition}, nil
	}
}

func chopContentType(contentType string) string {
	return strings.Split(contentType, ";")[0]
}

// Close cleans up underlying resources used by this client. The client is then unable to publish.
func (lc *StreamClient) Close() error {
	return lc.conn.Close()
}
