package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/itallix/go-metrics/internal/service"

	"github.com/go-resty/resty/v2"

	"github.com/itallix/go-metrics/internal/model"

	"github.com/itallix/go-metrics/internal/logger"
)

const (
	requestTimeoutSeconds = 10
)

type Sender interface {
	send(ctx context.Context) error
}

type Collector interface {
	collect()
}

type agent struct {
	runtime.MemStats

	Client            *resty.Client
	Counter           int64
	Gauges            map[string]model.Metrics
	RegisteredMetrics []string
	HashService       service.HashService
	RetryDelays       []time.Duration
}

func newAgent(client *resty.Client, secretKey string) *agent {
	var hashService service.HashService
	if secretKey != "" {
		hashService = service.NewHashService(secretKey)
	}
	return &agent{
		Client: client,
		Gauges: make(map[string]model.Metrics),
		RegisteredMetrics: []string{
			"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
			"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
			"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
		},
		HashService: hashService,
		RetryDelays: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (m *agent) collect() {
	runtime.ReadMemStats(&m.MemStats)

	m.Counter++

	var randomValue float64
	if err := binary.Read(rand.Reader, binary.BigEndian, &randomValue); err != nil {
		logger.Log().Errorf("error generating random number: %v", err)
	}
	m.Gauges["RandomValue"] = *model.NewGauge("RandomValue", &randomValue)

	for _, name := range m.RegisteredMetrics {
		v := reflect.ValueOf(m.MemStats)
		fieldVal := v.FieldByName(name)
		var vv float64

		if fieldVal.IsValid() {
			switch fieldVal.Kind() {
			case reflect.Uint, reflect.Uint64, reflect.Uint32:
				vv = float64(fieldVal.Uint())
			default:
				vv = fieldVal.Float()
			}
			m.Gauges[name] = *model.NewGauge(name, &vv)
		} else {
			logger.Log().Errorf("Field %s does not exist in MemStats\n", name)
		}
	}
}

func (m *agent) send(ctx context.Context) error {
	requestPath := "/updates"
	var metrics []model.Metrics
	for _, gauge := range m.Gauges {
		metrics = append(metrics, gauge)
	}
	metrics = append(metrics, *model.NewCounter("PollCount", &m.Counter))

	// Use encoder to pass autotests
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(metrics); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	c, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
	defer cancel()

	var (
		err  error
		resp *resty.Response
	)
	request := m.Client.R().
		SetContext(c).
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes())
	if m.HashService != nil {
		request.SetHeader(model.HashSha256Header, m.HashService.Sha256sum(buf.Bytes()))
	}

	for _, delay := range m.RetryDelays {
		resp, err = request.Post(requestPath)

		if err != nil {
			logger.Log().Errorf("Failed to send request, retrying after %v...", delay)
			time.Sleep(delay)
			continue
		}

		break
	}

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusOK {
		m.Counter = 0
	}

	return nil
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

	serverURL, config, err := parseFlags()
	if err != nil {
		logger.Log().Fatalf("Cannot parse flags: %v", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	defer tickerPoll.Stop()
	reportPoll := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer reportPoll.Stop()

	client := resty.New().SetBaseURL("http://"+serverURL.String()).
		SetHeader("Content-Type", "application/json")
	metricsAgent := newAgent(client, config.Key)

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				metricsAgent.collect()
				var metrics []model.Metrics
				for _, gauge := range metricsAgent.Gauges {
					metrics = append(metrics, gauge)
				}
				metrics = append(metrics, *model.NewCounter("PollCount", &metricsAgent.Counter))
				logger.Log().Infof("Collected metrics: %v", metrics)
			case <-reportPoll.C:
				logger.Log().Info("Sending metrics...")
				if err = metricsAgent.send(ctx); err != nil {
					logger.Log().Errorf("Error sending metrics %v", err)
				} else {
					logger.Log().Info("Metrics has been successfully sent")
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
