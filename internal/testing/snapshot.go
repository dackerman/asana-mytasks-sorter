package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SnapshotClient is an HTTP client that records and replays HTTP requests and responses
type SnapshotClient struct {
	// The real client to use when recording
	RealClient *http.Client
	
	// The directory to store snapshots
	SnapshotDir string
	
	// Whether to record or replay
	Mode string // "record" or "replay"
	
	// Recorded responses
	Snapshots map[string]SnapshotData
}

// SnapshotData represents a recorded HTTP request and response
type SnapshotData struct {
	Request struct {
		Method string            `json:"method"`
		URL    string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Body   string            `json:"body"`
	} `json:"request"`
	Response struct {
		StatusCode int               `json:"status_code"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
	} `json:"response"`
}

// NewSnapshotClient creates a new SnapshotClient
func NewSnapshotClient(realClient *http.Client, snapshotDir, mode string) (*SnapshotClient, error) {
	client := &SnapshotClient{
		RealClient:  realClient,
		SnapshotDir: snapshotDir,
		Mode:        mode,
		Snapshots:   make(map[string]SnapshotData),
	}
	
	// If in replay mode, load snapshots
	if mode == "replay" {
		err := client.loadSnapshots()
		if err != nil {
			return nil, err
		}
	}
	
	return client, nil
}

// loadSnapshots loads snapshots from disk
func (c *SnapshotClient) loadSnapshots() error {
	// Read all snapshot files
	files, err := os.ReadDir(c.SnapshotDir)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Read file
			path := filepath.Join(c.SnapshotDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			
			// Unmarshal JSON
			var snapshot SnapshotData
			err = json.Unmarshal(data, &snapshot)
			if err != nil {
				return err
			}
			
			// Add to map
			key := requestKey(snapshot.Request.Method, snapshot.Request.URL)
			c.Snapshots[key] = snapshot
		}
	}
	
	return nil
}

// saveSnapshot saves a snapshot to disk
func (c *SnapshotClient) saveSnapshot(key string, data SnapshotData) error {
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	// Generate a filename
	filename := strings.ReplaceAll(key, "/", "_") + ".json"
	path := filepath.Join(c.SnapshotDir, filename)
	
	// Write to file
	return os.WriteFile(path, jsonData, 0644)
}

// Do implements the http.Client interface
func (c *SnapshotClient) Do(req *http.Request) (*http.Response, error) {
	// Generate a key for this request
	key := requestKey(req.Method, req.URL.String())
	
	if c.Mode == "replay" {
		// Look up the response in our snapshots
		snapshot, ok := c.Snapshots[key]
		if !ok {
			return nil, fmt.Errorf("no snapshot found for request: %s", key)
		}
		
		// Create a response
		resp := &http.Response{
			StatusCode: snapshot.Response.StatusCode,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(snapshot.Response.Body)),
			Request:    req,
		}
		
		// Add headers
		for k, v := range snapshot.Response.Headers {
			resp.Header.Add(k, v)
		}
		
		return resp, nil
	}
	
	// Record mode - make a real request
	resp, err := c.RealClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	// Clone the response before we read the body
	var snapshot SnapshotData
	
	// Record request
	snapshot.Request.Method = req.Method
	snapshot.Request.URL = req.URL.String()
	snapshot.Request.Headers = make(map[string]string)
	for k, v := range req.Header {
		snapshot.Request.Headers[k] = strings.Join(v, ",")
	}
	
	// If the request has a body, try to read it
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		snapshot.Request.Body = string(bodyBytes)
		
		// Restore the request body for the real request
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	
	// Record response
	snapshot.Response.StatusCode = resp.StatusCode
	snapshot.Response.Headers = make(map[string]string)
	for k, v := range resp.Header {
		snapshot.Response.Headers[k] = strings.Join(v, ",")
	}
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	snapshot.Response.Body = string(bodyBytes)
	
	// Restore the response body
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	// Save the snapshot
	err = c.saveSnapshot(key, snapshot)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

// requestKey generates a unique key for a request
func requestKey(method, url string) string {
	return strings.ToLower(method) + "_" + cleanURL(url)
}

// cleanURL removes sensitive information from the URL
func cleanURL(url string) string {
	// Remove query parameters that might contain sensitive info
	parts := strings.Split(url, "?")
	return parts[0]
}

// NewSnapshotRoundTripper creates a new http.RoundTripper that records and replays HTTP interactions
func NewSnapshotRoundTripper(t *testing.T, snapshotDir string, mode string) http.RoundTripper {
	return &SnapshotRoundTripper{
		T:           t,
		SnapshotDir: snapshotDir,
		Mode:        mode,
		Snapshots:   make(map[string]SnapshotData),
		RealTransport: http.DefaultTransport,
	}
}

// SnapshotRoundTripper is an http.RoundTripper that records and replays HTTP interactions
type SnapshotRoundTripper struct {
	T           *testing.T
	SnapshotDir string
	Mode        string
	Snapshots   map[string]SnapshotData
	RealTransport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface
func (rt *SnapshotRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Generate a key for this request
	key := requestKey(req.Method, req.URL.String())
	
	// In replay mode, return the recorded response
	if rt.Mode == "replay" {
		// Load snapshots on first request
		if len(rt.Snapshots) == 0 {
			err := rt.loadSnapshots()
			if err != nil {
				rt.T.Fatalf("Failed to load snapshots: %v", err)
			}
		}
		
		// Look up the response in our snapshots
		snapshot, ok := rt.Snapshots[key]
		if !ok {
			rt.T.Fatalf("No snapshot found for request: %s", key)
		}
		
		// Create a response
		resp := &http.Response{
			StatusCode: snapshot.Response.StatusCode,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(snapshot.Response.Body)),
			Request:    req,
		}
		
		// Add headers
		for k, v := range snapshot.Response.Headers {
			resp.Header.Add(k, v)
		}
		
		return resp, nil
	}
	
	// In record mode, make a real request and record it
	resp, err := rt.RealTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	
	// Clone the response before we read the body
	var snapshot SnapshotData
	
	// Record request
	snapshot.Request.Method = req.Method
	snapshot.Request.URL = req.URL.String()
	snapshot.Request.Headers = make(map[string]string)
	for k, v := range req.Header {
		// Skip Authorization header to avoid recording the token
		if strings.ToLower(k) != "authorization" {
			snapshot.Request.Headers[k] = strings.Join(v, ",")
		}
	}
	
	// Record response
	snapshot.Response.StatusCode = resp.StatusCode
	snapshot.Response.Headers = make(map[string]string)
	for k, v := range resp.Header {
		snapshot.Response.Headers[k] = strings.Join(v, ",")
	}
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	snapshot.Response.Body = string(bodyBytes)
	
	// Restore the response body
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	// Save the snapshot
	err = rt.saveSnapshot(key, snapshot)
	if err != nil {
		rt.T.Logf("Failed to save snapshot: %v", err)
	}
	
	return resp, nil
}

// loadSnapshots loads snapshots from disk
func (rt *SnapshotRoundTripper) loadSnapshots() error {
	// Read all snapshot files
	files, err := os.ReadDir(rt.SnapshotDir)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Read file
			path := filepath.Join(rt.SnapshotDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			
			// Unmarshal JSON
			var snapshot SnapshotData
			err = json.Unmarshal(data, &snapshot)
			if err != nil {
				return err
			}
			
			// Add to map
			key := requestKey(snapshot.Request.Method, snapshot.Request.URL)
			rt.Snapshots[key] = snapshot
		}
	}
	
	return nil
}

// saveSnapshot saves a snapshot to disk
func (rt *SnapshotRoundTripper) saveSnapshot(key string, data SnapshotData) error {
	// Create snapshot directory if it doesn't exist
	err := os.MkdirAll(rt.SnapshotDir, 0755)
	if err != nil {
		return err
	}
	
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	// Generate a filename - remove characters that might be problematic in filenames
	filename := strings.ReplaceAll(key, "/", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, ".", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "=", "_")
	filename = strings.ReplaceAll(filename, "&", "_")
	filename += ".json"
	path := filepath.Join(rt.SnapshotDir, filename)
	
	// Write to file
	return os.WriteFile(path, jsonData, 0644)
}