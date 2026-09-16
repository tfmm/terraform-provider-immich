package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultTimeout bounds how long a single request to the Immich API may
// take. Without it, a hung or unreachable server would block
// terraform apply indefinitely. It is generous to accommodate asset
// uploads, which can be large; callers that need finer-grained control
// (e.g. cancellation) should rely on the context passed into each method.
const DefaultTimeout = 5 * time.Minute

type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

func NewClient(host, token string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
		HostURL:    host,
		Token:      token,
	}
}

// APIError represents a non-2xx response from the Immich API. It preserves
// the HTTP status code so callers can distinguish, for example, "not found"
// from other failures without parsing the error string.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.StatusCode, e.Body)
}

// IsNotFound reports whether err is an APIError with a 404 status code.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("x-api-key", c.Token)
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &APIError{StatusCode: res.StatusCode, Body: string(body)}
	}

	return body, nil
}
