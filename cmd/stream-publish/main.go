package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	client "github.com/projectriff/stream-client-go"
)

func main() {
	gateway := flag.String("gateway", "", "resolvable name of the streaming gateway")
	topic := flag.String("topic", "", "logical topic name")
	acceptableContentType := flag.String("accept", "*/*", "topic acceptable content type")
	contentType := flag.String("content-type", "", "payload content type")
	flag.Parse()

	streamClient, err := client.NewStreamClient(*gateway, *topic, *acceptableContentType)
	if err != nil {
		panic(err)
	}

	fmt.Println("Write payload and <ENTER>, <CTRL-C> to stop")
	reader := bufio.NewReader(os.Stdin)
	for {
		payload, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		publishResult, err := streamClient.Publish(context.TODO(), strings.NewReader(payload), nil, *contentType, map[string]string{})
		if err != nil {
			panic(err)
		}

		fmt.Printf("Result published at offset: %d, partition: %d\n", publishResult.Offset, publishResult.Partition)
	}
}
