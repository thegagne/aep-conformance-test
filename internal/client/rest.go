// Package client is a thin REST client used to exercise the API under test. It
// captures each request and response as evidence for conformance checks.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client issues requests against a base URL, applying optional auth.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	// Headers are applied to every request (e.g. an Authorization or API-key
	// header supplied via -H). Per-request headers and Auth take precedence.
	Headers map[string]string
	// Auth applies credentials to an outgoing request (bearer token, header…).
	Auth func(*http.Request)
}

// New returns a Client with the given base URL and per-request timeout.
func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: timeout},
	}
}

// Request records an issued request.
type Request struct {
	Method string
	URL    string
	Body   string
}

// Response is a captured HTTP response with the body pre-read and, when JSON,
// parsed into a generic structure.
type Response struct {
	Status  int
	Headers http.Header
	Body    string
	// JSON is the parsed body when the response is a JSON object; nil otherwise.
	JSON map[string]any
	// JSONArray is set when the top-level body is a JSON array.
	JSONArray []any
}

// IsJSON reports whether the response carried a JSON content type.
func (r *Response) IsJSON() bool {
	if r == nil {
		return false
	}
	return strings.Contains(r.Headers.Get("Content-Type"), "json")
}

// Do issues a request. path is a resource path (no leading slash needed) that
// may include a query string. body, if non-nil, is JSON-encoded. Optional
// header maps are applied to the request.
func (c *Client) Do(method, path string, body any, headers ...map[string]string) (Request, *Response, error) {
	url := c.BaseURL + "/" + strings.TrimLeft(path, "/")
	rec := Request{Method: method, URL: url}

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return rec, nil, fmt.Errorf("marshal body: %w", err)
		}
		rec.Body = string(b)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return rec, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	// Client-wide headers first, so per-request headers (e.g. If-Match) can
	// override them.
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	for _, hs := range headers {
		for k, v := range hs {
			req.Header.Set(k, v)
		}
	}
	if c.Auth != nil {
		c.Auth(req)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return rec, nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	out := &Response{Status: resp.StatusCode, Headers: resp.Header, Body: string(raw)}
	if len(bytes.TrimSpace(raw)) > 0 {
		trimmed := bytes.TrimSpace(raw)
		switch trimmed[0] {
		case '{':
			_ = json.Unmarshal(trimmed, &out.JSON)
		case '[':
			_ = json.Unmarshal(trimmed, &out.JSONArray)
		}
	}
	return rec, out, nil
}
