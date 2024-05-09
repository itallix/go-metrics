package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	pollIntervalSeconds   = 2
	reportIntervalSeconds = 10
	requestTimeoutSeconds = 10
	metricsServerURL      = "http://localhost:8080"
)

type Metrics struct {
	runtime.MemStats
	PollCount         uint64
	RandomValue       uint64
	Values            map[string]any
	RegisteredMetrics []string
}

func NewMetrics() *Metrics {
	return &Metrics{
		Values: make(map[string]any),
		RegisteredMetrics: []string{
			"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
			"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
			"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
		},
	}
}

func (m *Metrics) collect() {
	runtime.ReadMemStats(&m.MemStats)

	m.PollCount++

	var randomValue uint64
	err := binary.Read(rand.Reader, binary.BigEndian, &randomValue)
	if err != nil {
		log.Fatalf("error generating random number: %v", err)
	}
	m.RandomValue = randomValue

	typ := reflect.TypeOf(m.MemStats)
	val := reflect.ValueOf(m.MemStats)
	sTyp := reflect.TypeOf(m.MemStats.BySize)
	sVal := reflect.ValueOf(m.MemStats.BySize)
	for _, mm := range m.RegisteredMetrics {
		_, found := typ.FieldByName(mm)
		if !found {
			_, found = sTyp.FieldByName(mm)
			if !found {
				log.Printf("Metric '%s' was not found", mm)
			}
			m.Values[mm] = sVal.FieldByName(mm).Interface()
			continue
		}
		m.Values[mm] = val.FieldByName(mm).Interface()
	}
	m.Values["RandomValue"] = m.RandomValue
}

func (m *Metrics) send(ctx context.Context) {
	var wg sync.WaitGroup
	results := make(chan string, len(m.Values))

	for id, val := range m.Values {
		wg.Add(1)
		go func(id, val any) {
			defer wg.Done()
			metricsServerPath := fmt.Sprintf("/update/gauge/%s/%d", id, val)
			tctx, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(tctx, http.MethodPost, metricsServerURL+metricsServerPath, nil)
			if err != nil {
				results <- fmt.Sprintf("Cannot instantiate request object: %v", err)
				return
			}
			req.Header.Set("Content-Type", "text/plain")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- fmt.Sprintf("Error fetching %s: %v", id, err)
				return
			}
			results <- fmt.Sprintf("Success %s: %s", id, resp.Status)
			err = resp.Body.Close()
			if err != nil {
				log.Fatalf("Failed to close reponse body %v", err)
			}
		}(id, val)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		log.Println(result)
	}
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()

	tickerPoll := time.NewTicker(pollIntervalSeconds * time.Second)
	reportPoll := time.NewTicker(reportIntervalSeconds * time.Second)

	currentMetrics := NewMetrics()

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				currentMetrics.collect()
				log.Println(currentMetrics.Values)
			case <-reportPoll.C:
				log.Println("Sending metrics...")
				currentMetrics.send(ctx)
			}
		}
	}()

	<-quit
	log.Println("Shutting down gracefully")
}
