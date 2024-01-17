package traceloop

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/cenkalti/backoff"
)

func (sdk *Traceloop) GetVersion() string {
	info, ok := debug.ReadBuildInfo()
    if !ok {
        fmt.Printf("Failed to read build info")
        return ""
    }

	return info.Main.Version
}

func (sdk *Traceloop) fetchPath(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", sdk.Config.BaseURL, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sdk.Config.APIKey))
	req.Header.Set("X-Traceloop-SDK-Version", sdk.GetVersion())

	return sdk.Client.Do(req)
}

func (sdk *Traceloop) fetchPathWithRetry(path string, maxRetries uint64) (*http.Response, error) {
	var resp *http.Response
	
	err := backoff.Retry(func() error {
		var err error
		resp, err = sdk.fetchPath(path)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxRetries))

	return resp, err
}