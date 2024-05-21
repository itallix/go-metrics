package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/itallix/go-metrics/internal/service"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"

	"github.com/gin-gonic/gin"
)

type Result struct {
	Message string `json:"message"`
}

type MetricController struct {
	metricSrv service.MetricService
}

func NewMetricController(service service.MetricService) *MetricController {
	return &MetricController{
		metricSrv: service,
	}
}

func (mc *MetricController) UpdateMetric(c *gin.Context) {
	var metric model.Metrics
	if err := c.BindJSON(&metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "error decoding metric payload",
		})
		return
	}
	if err := mc.metricSrv.Update(&metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Log().Infof("Updating metric %s of type %s", metric.ID, metric.MType)
	c.JSON(http.StatusOK, metric)
}

func (mc *MetricController) GetMetric(c *gin.Context) {
	var metric model.Metrics
	if err := c.BindJSON(&metric); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "error decoding metric payload",
		})
		return
	}

	if err := mc.metricSrv.Read(&metric); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "metric is not found",
		})
		return
	}
	c.JSON(http.StatusOK, metric)
}

func (mc *MetricController) ListMetrics(c *gin.Context) {
	const tpl = `
<html>
<head>
	<title>Metrics</title>
</head>
<body>
	<ul>
		{{- range $key, $value := .Counters }}
		<li>{{ $key }}: {{ $value }}</li>
		{{- end }}
		{{- range $key, $value := .Gauges }}
		<li>{{ $key }}: {{ $value }}</li>
		{{- end }}
	</ul>
</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error rendering template")
		return
	}

	data := struct {
		Counters map[string]int64
		Gauges   map[string]float64
	}{
		Counters: mc.metricSrv.GetCounters(),
		Gauges:   mc.metricSrv.GetGauges(),
	}

	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)

	if err = t.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Error executing template")
	}
}

type MetricQuery struct {
	Type  model.MetricType `uri:"metricType,required"`
	ID    string           `uri:"metricName,required"`
	Value string           `uri:"metricValue,required"`
}

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

	if err := mc.metricSrv.Update(metric); err != nil {
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

	if err := mc.metricSrv.Read(metric); err != nil {
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
