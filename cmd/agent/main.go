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

	"github.com/go-resty/resty/v2"

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

	Client            *resty.Client
	Counter           int64
	Gauges            map[string]float64
	RegisteredMetrics []string
}

func newAgent(client *resty.Client) *agent {
	return &agent{
		Client: client,
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

func (m *agent) send(ctx context.Context) {
	// TODO: consider replacing with http client that supports retries
	var wg sync.WaitGroup
	results := make(chan string, len(m.Gauges))
	requestPath := "/update"

	for id, val := range m.Gauges {
		wg.Add(1)
		go func(id string, val float64) {
			defer wg.Done()
			tctx, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
			defer cancel()

			// Use encoder to pass autotests
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			if err := encoder.Encode(model.NewGauge(id, &val)); err != nil {
				logger.Log().Infof("Issue encoding gauge data to json: %v", err)
				return
			}
			resp, err := m.Client.R().
				SetContext(tctx).
				SetBody(&buf).
				Post(requestPath)

			if err != nil {
				results <- fmt.Sprintf("Issue sending update request to the server: %v", err)
				return
			}
			results <- fmt.Sprintf("Success %s: %d", id, resp.StatusCode())
		}(id, val)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// sending PollCount metric
		tctx, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
		defer cancel()

		resp, err := m.Client.R().
			SetContext(tctx).
			SetBody(model.NewCounter("PollCount", &m.Counter)).
			Post(requestPath)

		if err != nil {
			logger.Log().Infof("Issue sending PollCount to the server: %v", err)
			return
		}
		results <- fmt.Sprintf("Success %s: %d", "PollCount", resp.StatusCode())
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

	client := resty.New().SetBaseURL("http://"+serverURL.String()).
		SetHeader("Content-Type", "application/json")
	metricsAgent := newAgent(client)

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				metricsAgent.collect()
				logger.Log().Infof("Collected metrics: %v", metricsAgent.Gauges)
			case <-reportPoll.C:
				resp, httpErr := client.R().
					SetContext(ctx).
					Get("healthcheck")
				if httpErr != nil {
					logger.Log().Infof("Server is unavailable %v. Skip sending metrics...", httpErr)
				} else if resp.StatusCode() == http.StatusOK {
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
