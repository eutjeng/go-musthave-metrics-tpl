package config

import "time"

var (
	PollCount   int64
	RandomValue float64
)

const (
	ServerAddress  = "http://localhost:8080"
	PollInterval   = 2 * time.Second
	ReportInterval = 10 * time.Second
)
