package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client is a wrapper around Supabase REST API that handles authentication
// and request formatting. It supports both user-scoped requests (with JWT tokens)
// and admin-scoped requests (with service role key).
type Client struct {
	baseURL        string      // Supabase project URL (e.g., https://xxx.supabase.co)
	anonKey        string      // Public anon key for client requests
	serviceRoleKey string      // Service role key for admin operations (never expose to client)
	httpClient     *http.Client // HTTP client with timeout configuration
}

// NewClient creates a new Supabase client by reading credentials from environment variables.
// Required environment variables:
//   - SUPABASE_URL: Your Supabase project URL
//   - SUPABASE_ANON_KEY: Public anon key
//   - SUPABASE_SERVICE_ROLE_KEY: Service role key (admin access)
//
// Returns an error if any required environment variable is missing.
func NewClient() (*Client, error) {
	baseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	// Validate that all required credentials are present
	if baseURL == "" || anonKey == "" || serviceRoleKey == "" {
		return nil, fmt.Errorf("missing required Supabase environment variables")
	}

	return &Client{
		baseURL:        baseURL,
		anonKey:        anonKey,
		serviceRoleKey: serviceRoleKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Set reasonable timeout for all requests
		},
	}, nil
}

// Request makes an HTTP request to Supabase REST API with proper authentication.
// This is the core method that all other request methods use internally.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - method: HTTP method (GET, POST, PATCH, DELETE)
//   - path: API path (e.g., "/rest/v1/songs" or "/auth/v1/signup")
//   - body: Request body (will be JSON-encoded), can be nil for GET/DELETE
//   - token: User JWT token for authenticated requests, empty string for public
//   - useServiceRole: If true, uses service role key instead of user token (admin operations)
//
// Returns the HTTP response or an error if the request fails.
func (c *Client) Request(ctx context.Context, method, path string, body interface{}, token string, useServiceRole bool) (*http.Response, error) {
	var bodyReader io.Reader

	// Marshal body to JSON if present
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// Construct full URL by combining base URL with path
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required Supabase headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.anonKey) // Always required by Supabase

	// Set authorization header based on request type
	// Service role key bypasses RLS policies (admin operations only)
	// User token enforces RLS policies (normal user operations)
	if useServiceRole {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.serviceRoleKey))
	} else if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Get performs a GET request to Supabase with user authentication.
// Used for fetching data with Row Level Security (RLS) applied.
func (c *Client) Get(ctx context.Context, path string, token string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, path, nil, token, false)
}

// Post performs a POST request to Supabase with user authentication.
// Used for creating data with RLS applied (e.g., creating songs, comments).
func (c *Client) Post(ctx context.Context, path string, body interface{}, token string) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, body, token, false)
}

// Patch performs a PATCH request to Supabase with user authentication.
// Used for updating data with RLS applied (only owners can update).
func (c *Client) Patch(ctx context.Context, path string, body interface{}, token string) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, path, body, token, false)
}

// Delete performs a DELETE request to Supabase with user authentication.
// Used for deleting data with RLS applied (only owners can delete).
func (c *Client) Delete(ctx context.Context, path string, token string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, nil, token, false)
}

// ServiceRolePost performs a POST request using service role key (bypasses RLS).
// WARNING: This should only be used for admin operations.
// Use cases: Admin moderation, system-level operations.
func (c *Client) ServiceRolePost(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, body, "", true)
}

// ServiceRolePatch performs a PATCH request using service role key (bypasses RLS).
// WARNING: This should only be used for admin operations.
// Use cases: Admin user role updates, system-level modifications.
func (c *Client) ServiceRolePatch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, path, body, "", true)
}

// ServiceRoleDelete performs a DELETE request using service role key (bypasses RLS).
// WARNING: This should only be used for admin operations.
// Use cases: Admin content moderation, system-level deletions.
func (c *Client) ServiceRoleDelete(ctx context.Context, path string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, nil, "", true)
}

// ParseResponse is a utility function that parses HTTP response body into a struct.
// It handles error responses from Supabase by checking the status code and
// returns a SupabaseError if the request failed (status >= 400).
//
// Parameters:
//   - resp: HTTP response from Supabase
//   - v: Pointer to struct where response should be unmarshaled (can be nil)
//
// Returns error if response indicates failure or if JSON parsing fails.
func ParseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if Supabase returned an error (4xx or 5xx status codes)
	if resp.StatusCode >= 400 {
		return &SupabaseError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse successful response into provided struct
	if v != nil && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// SupabaseError represents an error response from Supabase API.
// This custom error type allows us to preserve both the HTTP status code
// and the error message from Supabase for better error handling.
type SupabaseError struct {
	StatusCode int    // HTTP status code (e.g., 400, 401, 403, 500)
	Message    string // Error message from Supabase
}

// Error implements the error interface for SupabaseError.
// Returns a formatted error message with status code and message.
func (e *SupabaseError) Error() string {
	return fmt.Sprintf("supabase error (status %d): %s", e.StatusCode, e.Message)
}

// IsSupabaseError checks if an error is of type SupabaseError.
// Useful for error handling to distinguish Supabase errors from other errors.
func IsSupabaseError(err error) bool {
	_, ok := err.(*SupabaseError)
	return ok
}
