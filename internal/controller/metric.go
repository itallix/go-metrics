package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/storage"
	"log"
	"net/http"
	"strconv"
)

type Result struct {
	Message string `json:"message"`
}

type MetricController struct {
	counters storage.Storage[int]
	gauges   storage.Storage[float64]
}

func NewMetricController(counters storage.Storage[int], gauges storage.Storage[float64]) *MetricController {
	return &MetricController{
		counters: counters,
		gauges:   gauges,
	}
}

func (mc *MetricController) UpdateMetric(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")
	metricValue := c.Param("metricValue")
	switch metricType {
	case "counter":
		metricValue, err := strconv.ParseInt(metricValue, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "metric type is not supported",
			})
			return
		}
		mc.counters.Set(metricName, int(metricValue))
	case "gauge":
		metricValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "metric type is not supported",
			})
			return
		}
		mc.gauges.Set(metricName, metricValue)
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "metric is not found",
		})
		return
	}

	log.Printf("Updating metric '%s' of type '%s' with value '%s'\n", metricName, metricType, metricValue)

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, Result{
		Message: "OK",
	})
}

func (mc *MetricController) GetMetric(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")

	switch metricType {
	case "counter":
		val, ok := mc.counters.Get(metricName)
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "metric is not found",
			})
			return
		}
		c.String(http.StatusOK, strconv.Itoa(val))
	case "gauge":
		val, ok := mc.gauges.Get(metricName)
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "metric is not found",
			})
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%g", val))
	default:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "metric is not found",
		})
	}
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
