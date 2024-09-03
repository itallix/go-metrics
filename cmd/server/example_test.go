//go:build example

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/itallix/go-metrics/internal/model"
)

func ExampleMain_updateBatch() {
	// Creating context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Creating collection of metrics to send to the server
	cv, gv := int64(1), 64.0
	metrics := []model.Metrics{*model.NewCounter("c0", &cv), *model.NewGauge("g0", &gv)}

	var buf bytes.Buffer
	// Creating gzip writer instance for payload compression
	gz := gzip.NewWriter(&buf)
	// Creating json encoder for converting the payload to compressed json bytes
	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(metrics); err != nil {
		fmt.Printf("error encoding metrics: %v", err)
	}
	if err := gz.Close(); err != nil {
		fmt.Printf("error closing gzip: %v", err)
	}

	// Using resty API as an example for sending metrics to the server.
	// Make sure you have your server running on localhost:8080, otherwise example will fail.
	client := resty.New().SetBaseURL("http://localhost:8080")

	// Sending updateBatch request
	resp, err := client.R().
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).SetContext(ctx).Post("updates/")

	if err != nil {
		fmt.Printf("error updating batch of metrics")
	}

	fmt.Println(resp.String())

	// Output:
	// [{"id":"c0","type":"counter","delta":1},{"id":"g0","type":"gauge","value":64}]
}
