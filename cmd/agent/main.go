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

	"github.com/itallix/go-metrics/internal/logger"
)

const (
	requestTimeoutSeconds = 10
)

type Sender interface {
	send(ctx context.Context, serverURL string)
}

type Collector interface {
	collect()
}

type agent struct {
	runtime.MemStats

	Counter           uint64
	Gauges            map[string]float64
	RegisteredMetrics []string
}

func newAgent() *agent {
	return &agent{
		Gauges: make(map[string]float64),
		RegisteredMetrics: []string{
			"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
			"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
			"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
		},
	}
}

func (m *agent) collect() {
	runtime.ReadMemStats(&m.MemStats)

	m.Counter++

	var randomValue float64
	if err := binary.Read(rand.Reader, binary.BigEndian, &randomValue); err != nil {
		logger.Log().Infof("error generating random number: %v", err)
	}
	m.Gauges["RandomValue"] = randomValue

	// TODO: consider refactor reflection part
	typ := reflect.TypeOf(m.MemStats)
	val := reflect.ValueOf(m.MemStats)
	sTyp := reflect.TypeOf(m.MemStats.BySize)
	sVal := reflect.ValueOf(m.MemStats.BySize)
	for _, mm := range m.RegisteredMetrics {
		_, found := typ.FieldByName(mm)
		if !found {
			_, found = sTyp.FieldByName(mm)
			if !found {
				logger.Log().Infof("Metric '%s' was not found", mm)
			}
			fieldVal := sVal.FieldByName(mm)
			switch fieldVal.Kind() {
			case reflect.Uint, reflect.Uint64, reflect.Uint32:
				m.Gauges[mm] = float64(fieldVal.Uint())
			default:
				m.Gauges[mm] = fieldVal.Float()
			}
			continue
		}
		fieldVal := val.FieldByName(mm)
		switch fieldVal.Kind() {
		case reflect.Uint, reflect.Uint64, reflect.Uint32:
			m.Gauges[mm] = float64(fieldVal.Uint())
		default:
			m.Gauges[mm] = fieldVal.Float()
		}
	}
}

func (m *agent) send(ctx context.Context, serverURL string) {
	// TODO: consider replacing with http client that supports retries
	var wg sync.WaitGroup
	results := make(chan string, len(m.Gauges))

	for id, val := range m.Gauges {
		wg.Add(1)
		go func(id string, val float64) {
			defer wg.Done()
			gaugeServerPath := fmt.Sprintf("/update/gauge/%s/%f", id, val)
			tctx, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(tctx, http.MethodPost, serverURL+gaugeServerPath, nil)
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
				logger.Log().Infof("Failed to close reponse body %v\n", err)
			}
		}(id, val)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// sending PollCount metric
		counterServerPath := fmt.Sprintf("/update/counter/PollCount/%d", m.Counter)
		resp, err := http.Post(serverURL+counterServerPath, "text/plain", nil)
		defer func() { _ = resp.Body.Close() }()
		if err != nil {
			logger.Log().Infof("Issue sending PollCount to the server: %v", err)
			return
		}
		m.Counter = 0
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		logger.Log().Info(result)
	}
}

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	defer func() {
		if deferErr := logger.Log().Sync(); deferErr != nil {
			logger.Log().Errorf("Failed to sync logger: %s", deferErr)
		}
	}()

	serverURL, intervalSettings, err := parseFlags()
	if err != nil {
		logger.Log().Fatalf("Cannot parse flags: %v", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickerPoll := time.NewTicker(intervalSettings.PollInterval)
	defer tickerPoll.Stop()
	reportPoll := time.NewTicker(intervalSettings.ReportInterval)
	defer reportPoll.Stop()

	metricsAgent := newAgent()

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				metricsAgent.collect()
				logger.Log().Infof("Collected metrics: %v", metricsAgent.Gauges)
			case <-reportPoll.C:
				url := "http://" + serverURL.String()
				resp, httpErr := http.Get(url + "/healthcheck")
				if httpErr != nil {
					logger.Log().Infof("Server is unavailable %v. Skip sending metrics...", httpErr)
				} else {
					_ = resp.Body.Close()
					logger.Log().Info("Sending metrics...")
					metricsAgent.send(ctx, url)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log().Info("Shutting down agent gracefully...")
	cancel()
}
