package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
	"github.com/dackerman/asana-tasks-sorter/internal/core"
	testing_util "github.com/dackerman/asana-tasks-sorter/internal/testing"
)

// TestMainWithSnapshots runs through the main workflow using recorded API responses
func TestMainWithSnapshots(t *testing.T) {
	// Determine whether to record or replay
	mode := "replay"
	if os.Getenv("RECORD") == "true" {
		mode = "record"
	}

	// Create a real client with the snapshot round tripper
	httpClient := &http.Client{
		Transport: testing_util.NewSnapshotRoundTripper(t, "snapshots", mode),
	}

	// Get the access token for recording mode
	var accessToken string
	if mode == "record" {
		accessToken = os.Getenv("ASANA_ACCESS_TOKEN")
		if accessToken == "" {
			t.Fatalf("ASANA_ACCESS_TOKEN environment variable is not set")
		}
	} else {
		// Use a dummy token for replay mode
		accessToken = "dummy_token"
	}

	// Create our Asana client with the snapshot client
	client := &asana.Client{
		Client:  httpClient,
		Token:   accessToken,
		BaseURL: asana.BaseURL,
	}

	// Use default configuration and dry run mode for tests
	config := asana.DefaultSectionConfig()
	dryRun := false

	// Run the core business logic
	_, err := core.OrganizeTasks(client, config, dryRun)
	if err != nil {
		t.Fatalf("Error in OrganizeTasks: %v", err)
	}
}