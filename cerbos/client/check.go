package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/reearth/reearthx/appx"
)

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
}

type GraphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

func CheckPermission(ctx context.Context, dashboardURL string, input CheckPermissionInput) (bool, error) {
	au := getAuthInfo(ctx)
	if au == nil {
		return false, fmt.Errorf("auth info not found")
	}

	query := `
		query CheckPermission($input: CheckPermissionInput!) {
			checkPermission(input: $input) {
				allowed
			}
		}
	`

	gqlRequest := GraphQLQuery{
		Query: query,
		Variables: map[string]interface{}{
			"input": input,
		},
	}

	requestBody, err := json.Marshal(gqlRequest)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", dashboardURL+"/api/graphql", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+au.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response CheckPermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data.CheckPermission.Allowed, nil
}

func getAuthInfo(ctx context.Context) *appx.AuthInfo {
	if v := ctx.Value("authinfo"); v != nil {
		if v2, ok := v.(appx.AuthInfo); ok {
			return &v2
		}
	}
	return nil
}
