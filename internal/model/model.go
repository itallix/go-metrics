package model

type Metrics struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
}

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

func NewCounter(id string, value *int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: Counter,
		Delta: value,
	}
}

func NewGauge(id string, value *float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: Gauge,
		Value: value,
	}
}
