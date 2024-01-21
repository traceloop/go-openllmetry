package traceloop

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/cenkalti/backoff"
)

func (instance *Traceloop) GetVersion() string {
	info, ok := debug.ReadBuildInfo()
    if !ok {
        fmt.Printf("Failed to read build info")
        return ""
    }

	return info.Main.Version
}

func (instance *Traceloop) fetchPath(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/%s", instance.config.BaseURL, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", instance.config.APIKey))
	req.Header.Set("X-Traceloop-SDK-Version", instance.GetVersion())

	return instance.Client.Do(req)
}

func (instance *Traceloop) fetchPathWithRetry(path string, maxRetries uint64) (*http.Response, error) {
	var resp *http.Response
	
	err := backoff.Retry(func() error {
		var err error
		resp, err = instance.fetchPath(path)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxRetries))

	return resp, err
}