package config

import "time"

type BackoffConfig struct {
	MaxRetries          uint64
}

type Config struct {
    BaseURL				string
    APIKey            	string
    PollingInterval   	time.Duration
    BackoffConfig     	BackoffConfig
}
