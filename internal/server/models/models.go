package models

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // metric value when type is 'gauge'
	Delta *int64   `json:"delta,omitempty"` // metric value when type is 'counter'
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter that takes the value 'gauge' or 'counter'
}
