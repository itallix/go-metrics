package main

import (
	"fmt"
	"log"
	"math/rand"
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
	metricsServerUrl      = "http://localhost:8080"
)

var registeredMetrics = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
}

type Metrics struct {
	runtime.MemStats
	PollCount   uint64
	RandomValue uint64
	values      map[string]any
}

func NewMetrics() *Metrics {
	return &Metrics{
		values: make(map[string]any),
	}
}

func (m *Metrics) collect() {
	runtime.ReadMemStats(&m.MemStats)

	m.PollCount++
	source := rand.NewSource(time.Now().UnixNano())
	m.RandomValue = rand.New(source).Uint64()

	typ := reflect.TypeOf(m.MemStats)
	val := reflect.ValueOf(m.MemStats)
	sTyp := reflect.TypeOf(m.MemStats.BySize)
	sVal := reflect.ValueOf(m.MemStats.BySize)
	for _, mm := range registeredMetrics {
		_, found := typ.FieldByName(mm)
		if !found {
			_, found = sTyp.FieldByName(mm)
			if !found {
				log.Printf("Metric '%s' was not found", mm)
			}
			m.values[mm] = sVal.FieldByName(mm).Interface()
			continue
		}
		m.values[mm] = val.FieldByName(mm).Interface()
	}
	m.values["RandomValue"] = m.RandomValue
}

func (m *Metrics) send() {
	var wg sync.WaitGroup
	results := make(chan string, len(m.values))

	for id, val := range m.values {
		wg.Add(1)
		go func(id, val any) {
			defer wg.Done()
			metricsServerPath := fmt.Sprintf("/update/gauge/%s/%d", id, val)
			resp, err := http.Post(metricsServerUrl+metricsServerPath, "text/plain", nil)
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
		fmt.Println(result)
	}
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	tickerPoll := time.NewTicker(pollIntervalSeconds * time.Second)
	reportPoll := time.NewTicker(reportIntervalSeconds * time.Second)

	currentMetrics := NewMetrics()

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				currentMetrics.collect()
				log.Println(currentMetrics.values)
			case <-reportPoll.C:
				log.Println("Sending metrics...")
				currentMetrics.send()
			}
		}
	}()

	<-quit
	log.Println("Shutting down gracefully")
}
