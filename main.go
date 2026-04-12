package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/grafpkg/server"
)

const (
	// Version is the current version of mcp-grafana.
	Version = "0.1.0"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %s, shutting down...", sig)
		cancel()
	}()

	if err := run(ctx); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

// run initializes and starts the MCP server with all Grafana tools registered.
func run(ctx context.Context) error {
	// Read configuration from environment variables
	// Default URL uses port 3000 (standard Grafana port).
	// Override with GRAFANA_URL env var to point at a remote or Docker instance.
	grafanaURL := getEnv("GRAFANA_URL", "http://localhost:3000")
	grafanaToken := os.Getenv("GRAFANA_API_KEY")
	// Default to stdio transport for use with Claude Desktop and VS Code extension.
	// Switch to MCP_TRANSPORT=sse for browser-based clients.
	// NOTE: I personally run this with SSE locally, so I've changed the default
	// transport to "sse" here. Set MCP_TRANSPORT=stdio to restore original behavior.
	transport := getEnv("MCP_TRANSPORT", "sse")

	log.Printf("Starting mcp-grafana %s", Version)
	log.Printf("Connecting to Grafana at %s", grafanaURL)
	log.Printf("Transport: %s | Token configured: %v", transport, grafanaToken != "")

	// Build server configuration
	cfg := server.Config{
		GrafanaURL:   grafanaURL,
		GrafanaToken: grafanaToken,
		Transport:    transport,
		Version:      Version,
	}

	// Create and configure the MCP server
	s, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Register all Grafana tools
	if err := tools.RegisterAll(s, cfg); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	log.Printf("MCP server starting with transport: %s", transport)

	// Start serving
	if err := s.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// getEnv returns the value of the environment variable named by key,
// or the fallback value if the variable is not set or empty.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
