package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/reearth/reearthx/appx"
)

const (
	checkPermissionQuery = `
        query CheckPermission($input: CheckPermissionInput!) {
            checkPermission(input: $input) {
                allowed
            }
        }
    `
	graphqlPath = "/api/graphql"
)

type Client struct {
	httpClient   *http.Client
	dashboardURL string
}

func NewClient(dashboardURL string) *Client {
	return &Client{
		httpClient:   &http.Client{},
		dashboardURL: dashboardURL,
	}
}

type CheckPermissionInput struct {
	Service  string `json:"service"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type CheckPermissionResponse struct {
	Data struct {
		CheckPermission struct {
			Allowed bool `json:"allowed"`
		} `json:"checkPermission"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type GraphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

func (c *Client) CheckPermission(ctx context.Context, authInfo *appx.AuthInfo, input CheckPermissionInput) (bool, error) {
	if err := c.validateInput(authInfo); err != nil {
		return false, err
	}

	req, err := c.createRequest(ctx, authInfo, input)
	if err != nil {
		return false, err
	}

	return c.executeRequest(req)
}

func (c *Client) validateInput(authInfo *appx.AuthInfo) error {
	if authInfo == nil {
		return fmt.Errorf("auth info is required")
	}
	return nil
}

func (c *Client) createRequest(ctx context.Context, authInfo *appx.AuthInfo, input CheckPermissionInput) (*http.Request, error) {
	gqlRequest := GraphQLQuery{
		Query: checkPermissionQuery,
		Variables: map[string]interface{}{
			"input": input,
		},
	}

	requestBody, err := json.Marshal(gqlRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.dashboardURL+graphqlPath, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, authInfo)
	return req, nil
}

func (c *Client) setHeaders(req *http.Request, authInfo *appx.AuthInfo) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authInfo.Token))
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) executeRequest(req *http.Request) (bool, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("server returned non-OK status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Response body: %s\n", string(bodyBytes))

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var response CheckPermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Errors) > 0 {
		return false, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Data.CheckPermission.Allowed {
		return true, nil
	}

	return false, nil
}
