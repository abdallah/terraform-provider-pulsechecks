package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

type Check struct {
	CheckID              string  `json:"check_id,omitempty"`
	Name                 string  `json:"name"`
	CheckType            string  `json:"check_type"`
	PeriodSeconds        int64   `json:"period_seconds"`
	GraceSeconds         int64   `json:"grace_seconds"`
	URL                  *string `json:"url,omitempty"`
	ExpectedStatusCode   *int64  `json:"expected_status_code,omitempty"`
	ExpectedString       *string `json:"expected_string,omitempty"`
	FailureThreshold     *int64  `json:"failure_threshold,omitempty"`
	Token                string  `json:"token,omitempty"`
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

func (c *Client) CreateCheck(ctx context.Context, teamID string, check Check) (*Check, error) {
	data, status, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/teams/%s/checks", teamID), check)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API error %d: %s", status, string(data))
	}
	var result Check
	return &result, json.Unmarshal(data, &result)
}

func (c *Client) GetCheck(ctx context.Context, teamID, checkID string) (*Check, error) {
	data, status, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/teams/%s/checks/%s", teamID, checkID), nil)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API error %d: %s", status, string(data))
	}
	var result Check
	return &result, json.Unmarshal(data, &result)
}

func (c *Client) UpdateCheck(ctx context.Context, teamID, checkID string, check Check) (*Check, error) {
	data, status, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/teams/%s/checks/%s", teamID, checkID), check)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API error %d: %s", status, string(data))
	}
	var result Check
	return &result, json.Unmarshal(data, &result)
}

func (c *Client) DeleteCheck(ctx context.Context, teamID, checkID string) error {
	data, status, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/teams/%s/checks/%s", teamID, checkID), nil)
	if err != nil {
		return err
	}
	if status == http.StatusNotFound {
		return nil
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("API error %d: %s", status, string(data))
	}
	return nil
}
