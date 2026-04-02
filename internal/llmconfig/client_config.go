package llmconfig

import "time"

type ClientConfig struct {
	Provider       string
	Endpoint       string
	APIKey         string
	Model          string
	Headers        map[string]string
	RequestTimeout time.Duration
}
