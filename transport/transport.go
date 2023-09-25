package transport

import (
	"context"
	"fmt"
	"net/http"
)

const (
	HeaderLastModified = "Last-Modified"
)

func GetResponse(ctx context.Context, cl *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("can't get feed. Error:%w", err)
	}

	response, err := cl.Do(req)
	if err != nil {
		return response, fmt.Errorf("can't get feed. Error:%w", err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("feed not availible. Status:%s", response.Status)
	}

	return response, err
}

func GetOnlyHeader(ctx context.Context, cl *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("can't get feed info. Error:%w", err)
	}

	response, err := cl.Do(req)
	if err != nil {
		return response, fmt.Errorf("can't get feed info. Error:%w", err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("feed not availible. Status:%s", response.Status)
	}

	return response, err
}
