// Package openapi provides OpenAPI spec fetching, parsing, and tool generation.
package openapi

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolDefinition represents a generated MCP tool from an OpenAPI operation.
type ToolDefinition struct {
	Name        string
	Description string
	HTTPMethod  string
	Path        string
	InputSchema map[string]any
	PathParams  []string
	QueryParams []string
	HasBody     bool
	Annotations *mcp.ToolAnnotations
}

// Generator generates MCP tool definitions from OpenAPI specifications.
type Generator struct{}

// NewGenerator creates a new tool generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateTools converts an OpenAPI spec into MCP tool definitions.
func (g *Generator) GenerateTools(spec *openapi3.T) []ToolDefinition {
	tools := []ToolDefinition{}

	if spec.Paths == nil {
		return tools
	}

	// Get the underlying map from Paths
	pathsMap := spec.Paths.Map()
	for path, pathItem := range pathsMap {
		if pathItem == nil {
			continue
		}

		sharedParams := g.extractParameters(pathItem.Parameters)

		// Get operations using Operations() method
		operations := pathItem.Operations()
		for method, operation := range operations {
			if operation == nil {
				continue
			}
			tool := g.buildToolDefinition(path, method, operation, sharedParams)
			if tool != nil {
				tools = append(tools, *tool)
			}
		}
	}

	return tools
}

// extractParameters extracts and resolves parameters from a PathItem.
func (g *Generator) extractParameters(params openapi3.Parameters) openapi3.Parameters {
	return params
}

// buildToolDefinition creates a ToolDefinition from an OpenAPI operation.
func (g *Generator) buildToolDefinition(path string, method string, operation *openapi3.Operation, sharedParams openapi3.Parameters) *ToolDefinition {
	if operation == nil {
		return nil
	}

	// Get operation ID
	operationId := operation.OperationID
	if operationId == "" {
		operationId = g.generateOperationId(path, method)
	}

	// Build description from summary and description
	description := g.buildDescription(operation)

	// Determine HTTP method
	httpMethod := strings.ToUpper(method)

	// Merge shared and operation parameters
	allParams := append(sharedParams, operation.Parameters...)

	// Extract parameters
	pathParams, queryParams, properties, required := g.extractParametersInfo(allParams)

	// Handle request body
	hasBody := false
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		hasBody = true
		bodyProps, bodyRequired := g.extractRequestBody(operation.RequestBody.Value)
		for k, v := range bodyProps {
			properties[k] = v
		}
		required = append(required, bodyRequired...)
	}

	// Build input schema
	inputSchema := map[string]any{
		"type": "object",
	}
	if len(properties) > 0 {
		inputSchema["properties"] = properties
	}
	if len(required) > 0 {
		inputSchema["required"] = uniqueStrings(required)
	}

	// Build annotations
	isDestructive := g.isDestructiveOperation(httpMethod, path, operationId)
	isMutatingGet := httpMethod == "GET" && isDestructive
	isReadOnly := httpMethod == "GET" && !isMutatingGet
	isIdempotent := contains([]string{"GET", "PUT", "DELETE"}, httpMethod) && !isMutatingGet

	// Get tag for prefix
	tag := ""
	if len(operation.Tags) > 0 {
		tag = operation.Tags[0]
	}

	// Add prefix to description
	prefix := ""
	if tag != "" {
		prefix = "[" + tag + "]"
	}
	if isDestructive {
		if prefix != "" {
			prefix += " "
		}
		prefix += "[DESTRUCTIVE]"
	}
	if prefix != "" {
		description = prefix + " " + description
	}

	// Truncate description to ~900 chars
	if len(description) > 900 {
		description = truncateString(description, 900)
	}

	// Build tool annotations
	// Note: DestructiveHint and OpenWorldHint are *bool, ReadOnlyHint and IdempotentHint are bool
	destructiveHint := isDestructive
	openWorldHint := true

	annotations := &mcp.ToolAnnotations{
		Title:           operation.Summary,
		ReadOnlyHint:    isReadOnly,
		DestructiveHint: &destructiveHint,
		IdempotentHint:  isIdempotent,
		OpenWorldHint:   &openWorldHint,
	}

	return &ToolDefinition{
		Name:        operationId,
		Description: description,
		HTTPMethod:  httpMethod,
		Path:        path,
		InputSchema: inputSchema,
		PathParams:  pathParams,
		QueryParams: queryParams,
		HasBody:     hasBody,
		Annotations: annotations,
	}
}

// buildDescription creates a combined description from summary and description.
func (g *Generator) buildDescription(operation *openapi3.Operation) string {
	summary := operation.Summary
	description := operation.Description

	result := summary
	if description != "" && description != summary {
		if result != "" {
			result += " — "
		}
		result += description
	}

	if result == "" {
		result = operation.OperationID
	}

	return strings.TrimSpace(result)
}

// extractParametersInfo extracts parameter information from a list of parameters.
func (g *Generator) extractParametersInfo(params openapi3.Parameters) (pathParams, queryParams []string, properties map[string]any, required []string) {
	pathParams = []string{}
	queryParams = []string{}
	properties = map[string]any{}
	required = []string{}

	for _, paramRef := range params {
		param := paramRef.Value
		if param == nil {
			continue
		}

		name := param.Name
		if name == "" {
			continue
		}

		// Build property definition
		propDef := map[string]any{
			"type": "string",
		}

		// Get schema from SchemaRef
		if param.Schema != nil && param.Schema.Value != nil {
			schema := param.Schema.Value
			// Type is *openapi3.Types, need to check differently
			if schema.Type != nil && len(*schema.Type) > 0 {
				propDef["type"] = (*schema.Type)[0]
			}
			if schema.Description != "" {
				propDef["description"] = schema.Description
			}
			if len(schema.Enum) > 0 {
				propDef["enum"] = schema.Enum
			}
			if schema.Format != "" {
				propDef["format"] = schema.Format
			}
		}

		if param.Description != "" {
			propDef["description"] = param.Description
		}

		in := param.In
		if in == "path" {
			pathParams = append(pathParams, name)
			required = append(required, name)
		} else if in == "query" {
			queryParams = append(queryParams, name)
			if param.Required {
				required = append(required, name)
			}
		} else {
			continue
		}

		properties[name] = propDef
	}

	return
}

// extractRequestBody extracts properties from the request body schema.
func (g *Generator) extractRequestBody(body *openapi3.RequestBody) (map[string]any, []string) {
	if body == nil {
		return nil, nil
	}

	props := map[string]any{}
	required := []string{}

	content := body.Content
	for mediaType, mediaTypeSchema := range content {
		if mediaType == "application/json" || mediaType == "multipart/form-data" {
			if mediaTypeSchema.Schema != nil && mediaTypeSchema.Schema.Value != nil {
				schema := mediaTypeSchema.Schema.Value
				if schema.Properties != nil {
					for propName, propSchemaRef := range schema.Properties {
						if propSchemaRef != nil && propSchemaRef.Value != nil {
							propDef := map[string]any{
								"type": "string",
							}
							s := propSchemaRef.Value
							// Type is *openapi3.Types
							if s.Type != nil && len(*s.Type) > 0 {
								propDef["type"] = (*s.Type)[0]
							}
							if s.Description != "" {
								propDef["description"] = s.Description
							}
							if len(s.Enum) > 0 {
								propDef["enum"] = s.Enum
							}
							props[propName] = propDef
						}
					}
				}
				if schema.Required != nil {
					required = schema.Required
				}
			}
			break
		}
	}

	return props, required
}

// generateOperationId generates an operation ID from path and method.
func (g *Generator) generateOperationId(path string, method string) string {
	parts := strings.Split(path, "/")
	parts = filterEmpty(parts)

	name := strings.ToUpper(method)
	for _, part := range parts {
		// Skip path parameters
		if strings.HasPrefix(part, "{") {
			continue
		}
		// Clean the part
		clean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '_'
		}, part)
		if clean != "" {
			name += "_" + clean
		}
	}

	return name
}

// isDestructiveOperation determines if an operation is destructive.
func (g *Generator) isDestructiveOperation(method, path, operationId string) bool {
	path = strings.ToLower(path)
	operationId = strings.ToLower(operationId)

	// DELETE methods are always destructive
	if method == "DELETE" {
		return true
	}

	// Check path for destructive terms
	destructivePathTerms := []string{
		"cancel", "delete", "refund", "purge", "wipe", "remove",
		"destroy", "reinstall", "reset_password", "change_root_password",
		"change_password", "mark_fraud", "disable", "suspend",
		"restore", "change_ip", "migration", "ipmi_power", "powerstrip",
		"null_routes", "clean_login_logs", "switch_port", "switchport_config",
		"mass_email", "buy_hd_space", "buy_ip",
	}

	for _, term := range destructivePathTerms {
		if strings.Contains(path, term) {
			if method == "POST" || method == "PUT" || method == "PATCH" {
				return true
			}
			if contains([]string{"reset_password", "change_password", "change_root_password", "reinstall", "restore", "ipmi_power", "powerstrip"}, term) {
				return true
			}
		}
	}

	// Check operationId patterns
	destructiveIdPatterns := []string{
		"cancel", "delete", "refund", "reassign", "suspend", "wipe", "purge", "remove",
		"resetpassword", "resetmailpassword", "reinstallos", "markfraud", "destroy",
		"restore", "migrate", "massemail", "apcpower", "apcsetup", "apcpowerstrip",
		"ipmipower", "changeip", "changepassword", "changerootpassword",
		"cleanloginlogs", "order", "manageswitchport", "addnullroute",
		"serveripmipower", "buyhdspace", "buyip", "forcedelete",
	}

	for _, pattern := range destructiveIdPatterns {
		if strings.HasPrefix(operationId, pattern) {
			return true
		}
	}

	return false
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func filterEmpty(slice []string) []string {
	result := []string{}
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func uniqueStrings(slice []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Try to find a good breaking point
	breakPoints := []string{". ", "? ", "! ", "\n\n"}
	cut := -1

	for _, bp := range breakPoints {
		idx := strings.LastIndex(s[:maxLen], bp)
		if idx > 700 { // At least 700 chars before break
			cut = idx + 1
			break
		}
	}

	if cut > 0 {
		return s[:cut]
	}

	// Fall back to hard truncation with ellipsis
	return s[:897] + "..."
}
