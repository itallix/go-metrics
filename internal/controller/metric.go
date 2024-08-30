package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage"
)

// Result used as a JSON response with a message field.
type Result struct {
	Message string `json:"message"`
}

// MetricController implements API handlers for metric requests.
type MetricController struct {
	metricsStorage storage.Storage
}

// NewMetricController constructs new controller instance with a storage for metrics.
func NewMetricController(metricsStorage storage.Storage) *MetricController {
	return &MetricController{
		metricsStorage: metricsStorage,
	}
}

// UpdateBatch updates a collection of metrics, expecting the payload in JSON format.
// The JSON payload should be an array of metric objects.
// It returns a 400 error if any of the metrics in the array do not conform to the expected model structure,
// and a 500 error if an error occurs during the save operation.
func (mc *MetricController) UpdateBatch(c *gin.Context) {
	var batch []model.Metrics
	if err := c.BindJSON(&batch); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "error decoding metric payload",
		})
		return
	}

	if err := mc.metricsStorage.UpdateBatch(c.Request.Context(), batch); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Log().Info("Updating metrics batch completed.")
	c.JSON(http.StatusOK, batch)
}

// UpdateOne updates a single metric, expecting the payload in JSON format.
// It returns a 400 error if the JSON payload's metric does not conform to the expected model structure,
// and a 500 error if an error occurs during the save operation.
func (mc *MetricController) UpdateOne(c *gin.Context) {
	var metric model.Metrics
	if err := c.BindJSON(&metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "error decoding metric payload",
		})
		return
	}
	if err := mc.metricsStorage.Update(c.Request.Context(), &metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Log().Infof("Updating metric %s of type %s", metric.ID, metric.MType)
	c.JSON(http.StatusOK, metric)
}

// GetMetric reads metric details from the storage.
// It returns 400 in case of incorrect JSON payload and 404 if request metric is not found.
func (mc *MetricController) GetMetric(c *gin.Context) {
	var metric model.Metrics
	if err := c.BindJSON(&metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "error decoding metric payload",
		})
		return
	}

	if err := mc.metricsStorage.Read(c.Request.Context(), &metric); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "metric is not found",
		})
		return
	}
	c.JSON(http.StatusOK, metric)
}

// ListMetrics returns HTML page with all registred metrics from storage.
// Returns 500 in case of errors when reading from storage.
func (mc *MetricController) ListMetrics(c *gin.Context) {
	const tpl = `
<html>
<head>
	<title>Metrics</title>
</head>
<body>
{{- $hasCounters := len .Counters }}
{{- $hasGauges := len .Gauges }}
{{- if or (gt $hasCounters 0) (gt $hasGauges 0) }}
	<ul>
		{{- range $key, $value := .Counters }}
		<li>{{ $key }}: {{ $value }}</li>
		{{- end }}
		{{- range $key, $value := .Gauges }}
		<li>{{ $key }}: {{ $value }}</li>
		{{- end }}
	</ul>
</body>
{{- else }}
<p>No metrics found</p>
{{- end }}
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error rendering template")
		return
	}

	ctx := c.Request.Context()
	counters, err := mc.metricsStorage.GetCounters(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading counters from storage")
		return
	}
	gauges, err := mc.metricsStorage.GetGauges(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading gauges from storage")
		return
	}

	data := struct {
		Counters map[string]int64
		Gauges   map[string]float64
	}{
		Counters: counters,
		Gauges:   gauges,
	}

	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)

	if err = t.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Error executing template")
	}
}

// MetricQuery used to describe query named parameters for legacy metrics requests.
type MetricQuery struct {
	Type  model.MetricType `uri:"metricType,required"`
	ID    string           `uri:"metricName,required"`
	Value string           `uri:"metricValue,required"`
}

// UpdatesMetricQuery updates a metric using query parameters.
// This is a legacy version of the handler that has been superseded by a JSON-based implementation.
func (mc *MetricController) UpdateMetricQuery(c *gin.Context) {
	var query MetricQuery
	if err := c.ShouldBindUri(&query); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "metric is not supported",
		})
		return
	}

	var metric *model.Metrics

	switch query.Type {
	case model.Counter:
		metricValue, err := strconv.ParseInt(query.Value, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "metric type is not supported",
			})
			return
		}
		metric = model.NewCounter(query.ID, &metricValue)
	case model.Gauge:
		metricValue, err := strconv.ParseFloat(query.Value, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "metric type is not supported",
			})
			return
		}
		metric = model.NewGauge(query.ID, &metricValue)
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "metric is not found",
		})
		return
	}

	if err := mc.metricsStorage.Update(c.Request.Context(), metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logger.Log().Infof("Updating metric '%s' of type '%s' with value '%s'", query.ID, query.Type, query.Value)

	c.JSON(http.StatusOK, Result{
		Message: "OK",
	})
}

// GetMetricQuery reads a metric from a storage using query parameters.
// This is a legacy version of the handler that has been superseded by a JSON-based implementation.
func (mc *MetricController) GetMetricQuery(c *gin.Context) {
	var query MetricQuery
	if err := c.BindUri(&query); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "metric is not supported",
		})
		return
	}

	var metric = &model.Metrics{
		ID:    query.ID,
		MType: query.Type,
	}

	if err := mc.metricsStorage.Read(c.Request.Context(), metric); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "metric is not found",
		})
		return
	}

	switch query.Type {
	case model.Counter:
		c.String(http.StatusOK, strconv.FormatInt(*metric.Delta, 10))
	case model.Gauge:
		c.String(http.StatusOK, fmt.Sprintf("%g", *metric.Value))
	default:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "metric is not found",
		})
	}
}
