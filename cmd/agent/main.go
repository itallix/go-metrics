package main

import (
	"bytes"
	"compress/gzip"
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

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/sync/errgroup"

	"github.com/itallix/go-metrics/internal/service"

	"github.com/go-resty/resty/v2"

	"github.com/itallix/go-metrics/internal/model"

	"github.com/itallix/go-metrics/internal/logger"
)

const (
	requestTimeoutSeconds = 10
)

var RuntimeMetrics = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
}

type Sender interface {
	send(ctx context.Context, jobs <-chan []model.Metrics, results chan<- error)
}

type Collector interface {
	collectRuntime() error
	collectExtra() error
}

type agent struct {
	runtime.MemStats

	Client      *resty.Client
	Counter     int64
	Gauges      map[string]model.Metrics
	HashService service.HashService
	RetryDelays []time.Duration
	mu          sync.RWMutex
}

func newAgent(client *resty.Client, secretKey string) *agent {
	var hashService service.HashService
	if secretKey != "" {
		hashService = service.NewHashService(secretKey)
	}
	return &agent{
		Client:      client,
		Gauges:      make(map[string]model.Metrics),
		HashService: hashService,
		RetryDelays: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (m *agent) collectRuntime() error {
	runtime.ReadMemStats(&m.MemStats)

	m.Counter++

	var randomValue float64
	if err := binary.Read(rand.Reader, binary.BigEndian, &randomValue); err != nil {
		return fmt.Errorf("error generating random number: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Gauges["RandomValue"] = *model.NewGauge("RandomValue", &randomValue)

	for _, name := range RuntimeMetrics {
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
			return fmt.Errorf("field %s does not exist in MemStats", name)
		}
	}

	return nil
}

func (m *agent) collectExtra() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("error collecting virtual memory: %w", err)
	}
	percentages, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return fmt.Errorf("error collecting cpu utilization: %w", err)
	}
	totalMem := float64(v.Total)
	freeMem := float64(v.Free)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Gauges["TotalMemory"] = *model.NewGauge("TotalMemory", &totalMem)
	m.Gauges["FreeMemory"] = *model.NewGauge("FreeMemory", &freeMem)
	m.Gauges["CPUutilization1"] = *model.NewGauge("CPUutilization1", &percentages[0])
	return err
}

func (m *agent) send(ctx context.Context, jobs <-chan []model.Metrics, results chan<- error) {
	for metrics := range jobs {
		logger.Log().Info("Processing job with batch of metrics")
		// Use encoder to pass autotests
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		encoder := json.NewEncoder(gz)
		if err := encoder.Encode(metrics); err != nil {
			results <- err
		}
		if err := gz.Close(); err != nil {
			results <- err
		}

		var (
			err  error
			resp *resty.Response
		)
		request := m.Client.R().
			SetHeader("Content-Encoding", "gzip").
			SetBody(buf.Bytes())
		if m.HashService != nil {
			request.SetHeader(model.HashSha256Header, m.HashService.Sha256sum(buf.Bytes()))
		}

		for _, delay := range m.RetryDelays {
			c, cancel := context.WithTimeout(ctx, requestTimeoutSeconds*time.Second)
			resp, err = request.SetContext(c).Post("updates/")
			cancel()
			if err != nil {
				logger.Log().Errorf("Failed to send request, retrying after %v...", delay)
				time.Sleep(delay)
				continue
			}

			break
		}

		if err != nil {
			results <- err
		}

		if resp.StatusCode() == http.StatusOK {
			m.Counter = 0
		}

		results <- nil
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

	serverURL, config, err := parseFlags()
	if err != nil {
		logger.Log().Fatalf("Cannot parse flags: %v", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan []model.Metrics, config.RateLimit)
	results := make(chan error, config.RateLimit)

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	defer tickerPoll.Stop()
	reportPoll := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer reportPoll.Stop()

	client := resty.New().SetBaseURL("http://"+serverURL.String()).
		SetHeader("Content-Type", "application/json")
	metricsAgent := newAgent(client, config.Key)

	for i := 0; i < config.RateLimit; i++ {
		go metricsAgent.send(ctx, jobs, results)
	}

	go func() {
		for err := range results {
			if err != nil {
				logger.Log().Errorf("Error sending metrics %v", err)
			} else {
				logger.Log().Info("Metrics has been successfully sent")
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				g := new(errgroup.Group)
				g.Go(metricsAgent.collectRuntime)
				g.Go(metricsAgent.collectExtra)

				if err = g.Wait(); err != nil {
					logger.Log().Errorf("Issue collecting metrics: %v", err)
				}

				var metrics []model.Metrics
				for _, gauge := range metricsAgent.Gauges {
					metrics = append(metrics, gauge)
				}
				metrics = append(metrics, *model.NewCounter("PollCount", &metricsAgent.Counter))
				logger.Log().Infof("Collected metrics: %v", metrics)
			case <-reportPoll.C:
				logger.Log().Info("Scheduling new job to send metrics...")
				var metrics []model.Metrics
				for _, gauge := range metricsAgent.Gauges {
					metrics = append(metrics, gauge)
				}
				metrics = append(metrics, *model.NewCounter("PollCount", &metricsAgent.Counter))
				jobs <- metrics
			case <-ctx.Done():
				return
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log().Info("Shutting down agent gracefully...")
	close(jobs)
	close(results)
	cancel()
}
