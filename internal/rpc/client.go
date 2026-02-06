package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	logging "github.com/ipfs/go-log/v2"

	appcfg "wallet-sign/internal/config"
)

var log = logging.Logger("rpc")

// Client represents a JSON-RPC client for communicating with Lotus API endpoints.
// It handles authentication, request formatting, and response parsing.
type Client struct {
	url    string
	token  string
	client *http.Client
}

// jsonRPCRequest represents a JSON-RPC 2.0 request structure.
type jsonRPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// jsonRPCResponse represents a JSON-RPC 2.0 response structure.
type jsonRPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// jsonRPCError represents a JSON-RPC 2.0 error object.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewLotusApi creates a new Lotus API client using environment variables for configuration.
// It reads LOTUS_API_URL (defaults to Glif public node) and LOTUS_API_TOKEN for authentication.
func NewLotusApi() *Client {
	log.Info("NewLotusApi: initializing Lotus API client")

	apiURL := appcfg.LotusConfig.Lotus.Host

	apiToken := appcfg.LotusConfig.Lotus.Token

	if apiToken != "" {
		log.Infof("NewLotusApi: connecting to %s (with token)", apiURL)
		fmt.Printf("Connecting to %s (with token)\n", apiURL)
	} else {
		log.Warnf("NewLotusApi: connecting to %s (no token)", apiURL)
		fmt.Printf("Connecting to %s (no token)\n", apiURL)
	}

	return &Client{
		url:    apiURL,
		token:  apiToken,
		client: &http.Client{},
	}
}

// Call executes a JSON-RPC method call on the Lotus API.
// The method name is automatically prefixed with "Filecoin.".
// If result is not nil, the response will be unmarshaled into it.
func (c *Client) Call(ctx context.Context, method string, params []interface{}, result interface{}) error {
	log.Debugf("Call: calling RPC method %s with %d params", method, len(params))

	reqBody := jsonRPCRequest{
		Jsonrpc: "2.0",
		Method:  "Filecoin." + method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Errorf("Call: failed to marshal request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("Call: failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Errorf("Call: failed to send request to %s: %v", c.url, err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Call: HTTP error %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Call: failed to read response: %v", err)
		return fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		log.Errorf("Call: failed to unmarshal response: %v", err)
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		log.Errorf("Call: RPC error for method %s: %s (code: %d)", method, rpcResp.Error.Message, rpcResp.Error.Code)
		return fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	if result != nil {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			log.Errorf("Call: failed to unmarshal result for method %s: %v", method, err)
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	log.Debugf("Call: successfully called %s", method)
	return nil
}
