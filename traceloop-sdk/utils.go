package traceloop

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cenkalti/backoff"
)

func (instance *Traceloop) fetchPath(path string) (*http.Response, error) {
	baseURL := instance.config.BaseURL
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	fullURL, err := url.JoinPath(baseURL, path)
	if err != nil {
		fmt.Printf("Failed to join URL path: %v\n", err)
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
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
