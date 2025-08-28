package traceloop

import "time"

type BackoffConfig struct {
	MaxRetries uint64
}

type Config struct {
	BaseURL         string
	APIKey          string
	Headers         map[string]string
	TracerName      string
	ServiceName     string
	PollingInterval time.Duration
	BackoffConfig   BackoffConfig
}
