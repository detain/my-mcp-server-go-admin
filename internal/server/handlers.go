// Package server provides MCP server implementation with transport detection.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/myadmin/go-mcp-proxy-admin/internal/openapi"
	"github.com/myadmin/go-mcp-proxy-admin/internal/proxy"
)

// ToolHandler handles MCP tool calls by proxying to the upstream API.
type ToolHandler struct {
	generator     *openapi.Generator
	proxyClient   *proxy.Client
	authExtractor *proxy.AuthExtractor
	toolDefs      []openapi.ToolDefinition
}

// NewToolHandler creates a new tool handler.
func NewToolHandler(specURL, apiBaseURL, bearerToken, apiKey, sessionID string) (*ToolHandler, error) {
	// Parse OpenAPI spec
	parser := openapi.NewParser("")
	spec, err := parser.FetchSpec(specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec: %w", err)
	}

	// Generate tools
	generator := openapi.NewGenerator()
	toolDefs := generator.GenerateTools(spec)

	slog.Info("generated tools from OpenAPI spec",
		"count", len(toolDefs),
		"spec_url", specURL,
	)

	// Create proxy client
	proxyClient := proxy.NewClient(apiBaseURL)

	// Create auth extractor
	authExtractor := proxy.NewAuthExtractor(bearerToken, apiKey, sessionID)

	return &ToolHandler{
		generator:     generator,
		proxyClient:   proxyClient,
		authExtractor: authExtractor,
		toolDefs:      toolDefs,
	}, nil
}

// ToolDefs returns the generated tool definitions.
func (h *ToolHandler) ToolDefs() []openapi.ToolDefinition {
	return h.toolDefs
}

// CreateToolHandler creates an MCP tool handler function for a given tool definition.
func (h *ToolHandler) CreateToolHandler(def openapi.ToolDefinition) func(context.Context, *mcp.CallToolRequest, any) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		slog.Debug("tool call received",
			"tool", def.Name,
			"method", def.HTTPMethod,
			"path", def.Path,
		)

		// Extract arguments from input
		args, ok := input.(map[string]any)
		if !ok {
			args = make(map[string]any)
		}

		// Get auth headers from request
		// Note: In HTTP mode, we need to get these from the HTTP request context
		// For now, use stdio auth
		authHeaders := h.authExtractor.Extract(nil)

		// Add MCP headers
		proxy.AddMCPHeaders(authHeaders, "")

		// Set auth headers on proxy client
		for key, value := range authHeaders {
			h.proxyClient.SetHeader(key, value)
		}

		// Build path parameters from arguments
		pathParams := make(map[string]string)
		for _, param := range def.PathParams {
			if val, ok := args[param]; ok {
				pathParams[param] = fmt.Sprintf("%v", val)
			}
		}

		// Build query parameters from arguments
		queryParams := make(map[string]string)
		for _, param := range def.QueryParams {
			if val, ok := args[param]; ok {
				queryParams[param] = fmt.Sprintf("%v", val)
			}
		}

		// Build body from remaining arguments
		var body map[string]any
		if def.HasBody {
			reserved := make(map[string]bool)
			for _, p := range def.PathParams {
				reserved[p] = true
			}
			for _, p := range def.QueryParams {
				reserved[p] = true
			}
			body = make(map[string]any)
			for key, val := range args {
				if !reserved[key] {
					body[key] = val
				}
			}
		}

		// Make request to upstream API
		proxyReq := proxy.Request{
			Method:      def.HTTPMethod,
			Path:        def.Path,
			PathParams:  pathParams,
			QueryParams: queryParams,
			Body:        body,
		}

		resp, err := h.proxyClient.Do(ctx, proxyReq)
		if err != nil {
			slog.Error("proxy request failed",
				"tool", def.Name,
				"error", err,
			)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: " + err.Error()},
				},
				IsError: true,
			}, nil, nil
		}

		// Handle error responses
		if resp.Error != "" {
			slog.Warn("proxy returned error",
				"tool", def.Name,
				"status", resp.StatusCode,
				"error", resp.Error,
			)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: resp.Error},
				},
				IsError: true,
			}, nil, nil
		}

		// Format response for MCP
		responseText, err := h.formatResponse(resp.Data)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error formatting response: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: responseText},
			},
		}, nil, nil
	}
}

// formatResponse formats the API response for MCP consumption.
func (h *ToolHandler) formatResponse(data any) (string, error) {
	if data == nil {
		return "", nil
	}

	// If it's a string, return it directly
	if str, ok := data.(string); ok {
		return str, nil
	}

	// Otherwise, marshal as JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonBytes), nil
}

// RegisterTools registers all generated tools with the MCP server.
func (h *ToolHandler) RegisterTools(server *Server) {
	for _, def := range h.toolDefs {
		tool := &mcp.Tool{
			Name:        def.Name,
			Description: def.Description,
		}

		// Build input schema
		inputSchema, err := h.buildInputSchema(def)
		if err != nil {
			slog.Warn("failed to build input schema for tool",
				"tool", def.Name,
				"error", err,
			)
			continue
		}
		tool.InputSchema = inputSchema

		handler := h.CreateToolHandler(def)

		// Register using generic AddTool
		mcp.AddTool(server.Impl(), tool, func(ctx context.Context, req *mcp.CallToolRequest, input map[string]any) (*mcp.CallToolResult, any, error) {
			return handler(ctx, req, input)
		})

		slog.Debug("registered tool",
			"name", def.Name,
			"method", def.HTTPMethod,
			"path", def.Path,
		)
	}
}

// buildInputSchema builds an MCP input schema from a tool definition.
func (h *ToolHandler) buildInputSchema(def openapi.ToolDefinition) (map[string]any, error) {
	// Convert the InputSchema map to the format expected by MCP
	schema := def.InputSchema
	if schema == nil {
		schema = map[string]any{"type": "object"}
	}
	return schema, nil
}
