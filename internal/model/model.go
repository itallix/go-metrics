package model

import "fmt"

// Metrics describes JSON payload for metrics of different types.
type Metrics struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
}

// MetricType is a string-based type reserved for metric types.
type MetricType string

const (
	Counter MetricType = "counter" // defines counter metric type with integer values
	Gauge   MetricType = "gauge"   // defines gauge metric type with float values
)

// NewCounter constructs new metric instance of type Counter with specified id and value.
func NewCounter(id string, value *int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: Counter,
		Delta: value,
	}
}

// NewGauge constructs new metric instance of type Gauge with specified id and value.
func NewGauge(id string, value *float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: Gauge,
		Value: value,
	}
}

// String gives a string representation of the metric instance.
// Example: "gauge: g01 = 2.345" or "counter: c01 = 64".
func (m Metrics) String() string {
	switch m.MType {
	case Gauge:
		return fmt.Sprintf("%s: %s = %f", Gauge, m.ID, *m.Value)
	case Counter:
		return fmt.Sprintf("%s: %s = %d", Counter, m.ID, *m.Delta)
	}
	return ""
}
