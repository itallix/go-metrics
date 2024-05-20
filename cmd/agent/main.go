package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
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

	"github.com/itallix/go-metrics/internal/model"

	"github.com/itallix/go-metrics/internal/logger"
)

const (
	requestTimeoutSeconds = 10
)

type Sender interface {
	send(ctx context.Context)
}

type Collector interface {
	collect()
}

type agent struct {
	runtime.MemStats

	ServerURL         string
	Counter           int64
	Gauges            map[string]float64
	RegisteredMetrics []string
}

func newAgent(serverURL string) *agent {
	return &agent{
		ServerURL: serverURL,
		Gauges:    make(map[string]float64),
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

func (m *agent) send(ctx context.Context) {
	// TODO: consider replacing with http client that supports retries
	var wg sync.WaitGroup
	results := make(chan string, len(m.Gauges))
	requestURL := m.ServerURL + "/update"

	for id, val := range m.Gauges {
		wg.Add(1)
		go func(id string, val float64) {
			defer wg.Done()
			tctx, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
			defer cancel()

			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			if err := encoder.Encode(model.NewGauge(id, &val)); err != nil {
				logger.Log().Infof("Issue encoding gauge data to json: %v", err)
				return
			}
			req, err := http.NewRequestWithContext(tctx, http.MethodPost, requestURL, &buf)
			if err != nil {
				results <- fmt.Sprintf("Cannot instantiate request object: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
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
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(model.NewCounter("PollCount", &m.Counter)); err != nil {
			logger.Log().Infof("Issue encoding counter data to json: %v", err)
			return
		}
		resp, err := http.Post(requestURL, "application/json", &buf)
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

	tickerPoll := time.NewTicker(intervalSettings.PollInterval)
	defer tickerPoll.Stop()
	reportPoll := time.NewTicker(intervalSettings.ReportInterval)
	defer reportPoll.Stop()

	httpServerURL := "http://" + serverURL.String()
	metricsAgent := newAgent(httpServerURL)

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				metricsAgent.collect()
				logger.Log().Infof("Collected metrics: %v", metricsAgent.Gauges)
			case <-reportPoll.C:
				resp, httpErr := http.Get(httpServerURL)
				if httpErr != nil {
					logger.Log().Infof("Server is unavailable %v. Skip sending metrics...", httpErr)
				} else {
					_ = resp.Body.Close()
					logger.Log().Info("Sending metrics...")
					metricsAgent.send(ctx)
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
