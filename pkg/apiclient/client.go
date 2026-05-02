package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	BaseURL      string
	AuthURL      string
	ClientID     string
	ClientSecret string
	CliConfig    *CLIUserConfig
	Token        string
	HTTPClient   *http.Client
}

// GraphQLRequest represents a GraphQL request payload
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response structure
type GraphQLResponse struct {
	Data       map[string]interface{} `json:"data"`
	Errors     []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []string               `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ErrorResponse represents the structure of error responses from the API
type ErrorResponse struct {
	Type     string              `json:"type,omitempty"`
	Title    string              `json:"title,omitempty"`
	Status   int                 `json:"status,omitempty"`
	Errors   map[string][]string `json:"errors,omitempty"`
	TraceID  string              `json:"traceId,omitempty"`
	Detail   string              `json:"detail,omitempty"`
	Instance string              `json:"instance,omitempty"`
}

func NewClientFromCLI(configPath string) (*Client, error) {
	return NewClientFromCLIWithEndpoint(configPath, "")
}

func NewClientFromCLIWithEndpoint(configPath string, endpointName string) (*Client, error) {

	if configPath == "" {
		foundConfigPath, err := findCLIUserConfigPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get config path: %v", err)
		}

		configPath = foundConfigPath
	}

	return NewClientWithAllConfigAndEndpoint("", "", "", "", configPath, endpointName, &http.Client{})
}

func NewClientFromEnv() (*Client, error) {
	clientID := os.Getenv("SITECORE_AUTHORING_CLIENT_ID")
	clientSecret := os.Getenv("SITECORE_AUTHORING_CLIENT_SECRET")

	return NewClientWithAllConfig("", "", clientID, clientSecret, "", &http.Client{})
}

func NewClient(clientID string, clientSecret string) (*Client, error) {
	return NewClientWithAllConfig("", "", clientID, clientSecret, "", &http.Client{})
}

func NewClientWithHost(host string, clientID string, clientSecret string) (*Client, error) {
	return NewClientWithAllConfig(host, "", clientID, clientSecret, "", &http.Client{})
}

func NewClientWithAllConfig(baseUrl string, authUrl string, clientId string, clientSecret string, cliUserConfigPath string, httpClient *http.Client) (*Client, error) {
	return NewClientWithAllConfigAndEndpoint(baseUrl, authUrl, clientId, clientSecret, cliUserConfigPath, "", httpClient)
}

func NewClientWithAllConfigAndEndpoint(baseUrl string, authUrl string, clientId string, clientSecret string, cliUserConfigPath string, endpointName string, httpClient *http.Client) (*Client, error) {
	BaseURL := ""
	AuthURL := "https://auth.sitecorecloud.io"

	if len(baseUrl) > 0 {
		BaseURL = baseUrl
	}

	if len(authUrl) > 0 {
		AuthURL = authUrl
	}

	var cliConfig *CLIUserConfig

	if cliUserConfigPath != "" {
		cfg, err := readCLIUserConfig(cliUserConfigPath)
		if cfg == nil || err != nil {
			return nil, fmt.Errorf("failed to read specified cli config from %s: %v", cliUserConfigPath, err)
		}

		cliConfig = cfg

		// If urls are not explicitly overriden, then use values from config
		if len(baseUrl) == 0 {
			// Use the specified endpoint's host, or first endpoint if not specified
			if endpointName != "" {
				if endpoint, exists := cliConfig.Endpoints[endpointName]; exists {
					BaseURL = endpoint.Host
				} else {
					return nil, fmt.Errorf("endpoint '%s' not found in CLI config", endpointName)
				}
			} else {
				// Use the first endpoint's host as default
				for _, endpoint := range cliConfig.Endpoints {
					BaseURL = endpoint.Host
					break
				}
			}
		}

		if len(authUrl) == 0 {
			// Use the specified endpoint's authority, or first endpoint if not specified
			if endpointName != "" {
				if endpoint, exists := cliConfig.Endpoints[endpointName]; exists {
					AuthURL = endpoint.Authority
				} else {
					return nil, fmt.Errorf("endpoint '%s' not found in CLI config", endpointName)
				}
			} else {
				// Use the first endpoint's authority as default
				for _, endpoint := range cliConfig.Endpoints {
					AuthURL = endpoint.Authority
					break
				}
			}
		}
	}

	if cliConfig == nil && (len(clientId) == 0 || len(clientSecret) == 0) {
		return nil, fmt.Errorf("client_id and client_secret must be provided")
	}

	// If not using CLI config, base URL must be provided
	if cliConfig == nil && len(BaseURL) == 0 {
		return nil, fmt.Errorf("base_url must be provided when not using CLI authentication")
	}

	return &Client{
		BaseURL:      strings.TrimSuffix(BaseURL, "/"),
		AuthURL:      strings.TrimSuffix(AuthURL, "/"),
		ClientID:     clientId,
		ClientSecret: clientSecret,
		CliConfig:    cliConfig,
		HTTPClient:   httpClient,
	}, nil
}

// doGraphQLRequest executes a GraphQL query/mutation
type GraphQLRequestOptions struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
}

func (c *Client) doGraphQLRequest(opts GraphQLRequestOptions) (*GraphQLResponse, error) {
	// Ensure we have a valid token
	err := c.EnsureTokenValid()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure valid token: %v", err)
	}

	// Create request URL - for Sitecore Authoring API, we need to use the graphql endpoint
	requestURL := fmt.Sprintf("%s/sitecore/api/authoring/graphql/v1/", c.BaseURL)

	// Create GraphQL request payload
	requestBody := GraphQLRequest{
		Query:     opts.Query,
		Variables: opts.Variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request body: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	// Set custom headers if provided
	if opts.Headers != nil {
		for key, value := range opts.Headers {
			req.Header.Set(key, value)
		}
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send GraphQL request: %v", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GraphQL request to %s failed with status code %d: %s", requestURL, resp.StatusCode, string(body))
	}

	// Parse response
	var graphQLResponse GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to decode GraphQL response: %v", err)
	}

	// Check for GraphQL errors
	if len(graphQLResponse.Errors) > 0 {
		errorMessages := []string{}
		for _, err := range graphQLResponse.Errors {
			errorMessages = append(errorMessages, err.Message)
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(errorMessages, ", "))
	}

	return &graphQLResponse, nil
}
