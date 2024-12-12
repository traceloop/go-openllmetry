package traceloop

import (
	"fmt"
	"net/http"

	"github.com/cenkalti/backoff"
)

func (instance *Traceloop) fetchPath(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/%s", instance.config.BaseURL, path), nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", instance.config.APIKey))
	req.Header.Set("X-Traceloop-SDK-Version", Version())

	return instance.Client.Do(req)
}

func (instance *Traceloop) fetchPathWithRetry(path string, maxRetries uint64) (*http.Response, error) {
	var resp *http.Response

	err := backoff.Retry(func() error {
		var err error
		resp, err = instance.fetchPath(path)
		if err != nil {
			fmt.Printf("Failed to fetch path: %v\n", err)
		}
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxRetries))

	return resp, err
}
