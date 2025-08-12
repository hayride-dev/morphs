package ory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hayride-dev/bindings/go/hayride/mcp/auth"
	"github.com/hayride-dev/bindings/go/wasi/net/http/transport"
)

var _ auth.Provider = (*oryProvider)(nil)

type oryProvider struct {
	registerEndpoint   string
	authEndpoint       string
	tokenEndpoint      string
	introspectEndpoint string

	contact string
	apiKey  string

	// Map of generated codes to ory tokens
	client *http.Client
}

// Return the URL used for authorization
func (n *oryProvider) AuthURL() (string, error) {
	return n.authEndpoint, nil
}

// Register a new client with the provider
func (n *oryProvider) Registration(data []byte) ([]byte, error) {
	clientIdGeneration := false

	// Parse json data into client info map
	clientInfo := map[string]interface{}{}
	if err := json.Unmarshal(data, &clientInfo); err != nil {
		return nil, fmt.Errorf("failed to parse client info: %w", err)
	}

	// Add additional required fields
	clientInfo["contacts"] = []string{
		n.contact,
	}
	if clientIdGeneration {
		clientId := uuid.New().String()
		clientInfo["client_id"] = clientId
		clientInfo["client_id_issued_at"] = time.Now().Unix()
	}

	// Post with json-encoded clientInfo to the registration URL
	clientInfoJSON, err := json.Marshal(clientInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal client info: %w", err)
	}

	fmt.Printf("Sending registration request body: %s\n", string(clientInfoJSON))

	req, err := http.NewRequest("POST", n.registerEndpoint, bytes.NewReader(clientInfoJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Return body as a string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}
	return body, nil
}

// Exchange the code for an access token (returns the token json response as a string)
func (n *oryProvider) ExchangeCode(data []byte) ([]byte, error) {
	// Make our own HTTP Request to get the Token
	req, err := http.NewRequest("POST", n.tokenEndpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Return the body as a string
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}
	return response, nil
}

type introspectionResponse struct {
	Active   bool   `json:"active"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
	Exp      int64  `json:"exp"` // expiration timestamp
}

// Validate Token through introspection
// Reference: https://github.com/ory/mcp/blob/main/packages/mcp-oauth-provider/src/index.ts#L351
func (n *oryProvider) Validate(token string) (bool, error) {
	// Build URL-encoded form body
	form := url.Values{}
	form.Set("token", token)
	form.Set("token_type_hint", "access_token")

	req, err := http.NewRequest("POST", n.introspectEndpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+n.apiKey)

	// Send request
	resp, err := n.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Decode JSON body
	var result introspectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed decoding response: %w", err)
	}

	return result.Active, nil
}

func Constructor() (auth.Provider, error) {
	client := &http.Client{
		Transport: transport.New(),
	}

	baseURL := os.Getenv("ORY_API_URL")
	if baseURL == "" {
		fmt.Println("ORY_API_URL environment variable is not set")
		return nil, fmt.Errorf("ORY_API_URL environment variable is not set")
	}

	apiKey := os.Getenv("ORY_API_KEY")
	if apiKey == "" {
		fmt.Println("ORY_API_KEY environment variable is not set")
		return nil, fmt.Errorf("ORY_API_KEY environment variable is not set")
	}

	return &oryProvider{
		registerEndpoint:   baseURL + "/oauth2/register",
		authEndpoint:       baseURL + "/oauth2/auth",
		tokenEndpoint:      baseURL + "/oauth2/token",
		introspectEndpoint: baseURL + "/admin/oauth2/introspect",
		contact:            "support@hayride.dev",
		apiKey:             apiKey,
		client:             client,
	}, nil
}
