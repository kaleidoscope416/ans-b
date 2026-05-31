package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *HTTPClient) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if c == nil || c.baseURL == "" {
		return nil, fmt.Errorf("embedding service url is not configured")
	}
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts is required")
	}

	body, err := json.Marshal(struct {
		Texts []string `json:"texts"`
	}{Texts: texts})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("embedding service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var result struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(result.Embeddings), len(texts))
	}
	for i, vector := range result.Embeddings {
		if len(vector) == 0 {
			return nil, fmt.Errorf("embedding %d is empty", i)
		}
	}
	return result.Embeddings, nil
}

func VectorLiteral(vector []float64) string {
	parts := make([]string, len(vector))
	for i, value := range vector {
		parts[i] = strconv.FormatFloat(value, 'f', 8, 64)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
