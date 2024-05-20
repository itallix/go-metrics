package controller

import (
	"fmt"
	"net/http"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/storage"
)

type Result struct {
	Message string `json:"message"`
}

type MetricController struct {
	counters storage.Storage[int64]
	gauges   storage.Storage[float64]
}

func NewMetricController(counters storage.Storage[int64], gauges storage.Storage[float64]) *MetricController {
	return &MetricController{
		counters: counters,
		gauges:   gauges,
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
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "metric type is not supported",
			})
			return
		}
		val := mc.counters.Update(metric.ID, *metric.Delta)
		metric.Delta = &val
	case model.Gauge:
		if metric.Value == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "metric type is not supported",
			})
			return
		}
		val := mc.gauges.Set(metric.ID, *metric.Value)
		metric.Value = &val
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "metric is not found",
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

	switch metric.MType {
	case model.Counter:
		val, ok := mc.counters.Get(metric.ID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "metric is not found",
			})
			return
		}
		metric.Delta = &val
	case model.Gauge:
		val, ok := mc.gauges.Get(metric.ID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "metric is not found",
			})
			return
		}
		metric.Value = &val
	default:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "metric is not found",
		})
		return
	}
	c.JSON(http.StatusOK, metric)
}

func (mc *MetricController) ListMetrics(c *gin.Context) {
	_, _ = c.Writer.WriteString(fmt.Sprintln("<html><head><title>Metrics</title></head><body><ul>"))

	for k, v := range mc.counters.Copy() {
		_, _ = c.Writer.WriteString(fmt.Sprintf("<li>%s: %d</li>", k, v))
	}

	for k, v := range mc.gauges.Copy() {
		_, _ = c.Writer.WriteString(fmt.Sprintf("<li>%s: %f</li>", k, v))
	}

	_, _ = c.Writer.WriteString(fmt.Sprintln("</ul></body></html>"))
	c.Status(http.StatusOK)
}
