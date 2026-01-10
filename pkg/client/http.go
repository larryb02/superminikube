package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"superminikube/pkg/apiserver/watch"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	// This is components node identifier
	nodeName string
}

func NewHTTPClient(baseURL, nodeName string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		nodeName:   nodeName,
	}
}

func (c *HTTPClient) Get(ctx context.Context, resource string, id uuid.UUID) ([]byte, error) {
	// TODO: Refactor url, all url's besides watch may be broken at the moment...
	url := fmt.Sprintf("%s/api/v1/%s/%s", c.baseURL, resource, id) // resource?nodename/id?
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", resource, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

func (c *HTTPClient) List(ctx context.Context, resource string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/%s", c.baseURL, resource)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %v", resource, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

func (c *HTTPClient) Update(ctx context.Context, resource string, id uuid.UUID, data []byte) error {
	url := fmt.Sprintf("%s/api/v1/%s/%s", c.baseURL, resource, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update %s: %v", resource, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *HTTPClient) Watch(ctx context.Context) (<-chan watch.WatchEvent, error) {
	eventChan := make(chan watch.WatchEvent)
	defaultDelay := 1
	const maxAttempts = 3
	go func() {
		defer close(eventChan)

		for {
			select {
			case <-ctx.Done():
				slog.Debug("cancelled watch context")
				return
			default:
				err := WithRetry(ctx, RetryOptions{
					Attempts: maxAttempts,
					Delay:    time.Duration(defaultDelay) * time.Second,
					Backoff: func(attempt int) time.Duration {
						newDelay := defaultDelay * 2
						defaultDelay = newDelay
						slog.Debug("new delay", "timer", newDelay)
						return time.Second * time.Duration(newDelay)
					},
					ShouldRetry: func(err error) bool {
						return errors.Is(err, context.DeadlineExceeded) ||
							strings.Contains(err.Error(), "connection refused")
					},
				},
					func(ctx context.Context) error {
						return c.watchStream(ctx, eventChan)
					})
				if err != nil {
					slog.Error("watch stream error", "error", err)
					return
				}
			}
		}
	}()

	return eventChan, nil
}

func parseStream(line string) (watch.WatchEvent, error) {
	// Ignore comments/keepalives
	if strings.HasPrefix(line, ":") {
		return watch.WatchEvent{}, fmt.Errorf("keepalive")
	}

	if line == "" {
		return watch.WatchEvent{}, fmt.Errorf("empty line")
	}

	if strings.HasPrefix(line, "data: ") {
		data := strings.TrimPrefix(line, "data: ")
		var event watch.WatchEvent
		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			return watch.WatchEvent{}, err
		}
		return event, nil
	}

	return watch.WatchEvent{}, fmt.Errorf("unknown line format")
}

func (c *HTTPClient) watchStream(ctx context.Context, eventChan chan<- watch.WatchEvent) error {
	url := fmt.Sprintf("%s/api/v1/watch?nodename=%s", c.baseURL, c.nodeName)
	slog.Debug(fmt.Sprintf("making request to %s", url))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Use a client with no timeout for SSE long-lived connection
	watchClient := &http.Client{Timeout: 0}
	resp, err := watchClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to watch stream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			line := scanner.Text()
			slog.Debug("received SSE event", "line", line)
			parsedEvent, err := parseStream(line)
			if err != nil {
				slog.Debug("nothing to do.")
				continue
			}
			eventChan <- parsedEvent
		}
	}

	return scanner.Err()
}

func (c *HTTPClient) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/healthz", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping apiserver: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("apiserver health check failed: status %d", resp.StatusCode)
	}

	return nil
}
