package models

type Metrics struct {
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter that takes the value 'gauge' or 'counter'
	Delta *int64   `json:"delta,omitempty"` // metric value when type is 'counter'
	Value *float64 `json:"value,omitempty"` // metric value when type is 'gauge'
}

type MetricsQuery struct {
	ID    string `json:"id"`   // metric name
	MType string `json:"type"` // parameter that takes the value 'gauge' or 'counter'
}
