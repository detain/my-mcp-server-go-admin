package openapi

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestGenerateToolsEmptySpec(t *testing.T) {
	g := NewGenerator()
	spec := &openapi3.T{}
	tools := g.GenerateTools(spec)
	if len(tools) != 0 {
		t.Errorf("expected 0 tools for empty spec, got %d", len(tools))
	}
}

func TestGenerateToolsWithOperation(t *testing.T) {
	g := NewGenerator()

	// Create paths using NewPaths and Set methods
	paths := openapi3.NewPaths()
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "getUsers",
			Summary:     "List users",
			Description: "Returns a list of all users",
		},
		Post: &openapi3.Operation{
			OperationID: "createUser",
			Summary:     "Create user",
			Description: "Creates a new user",
		},
	}
	paths.Set("/users", pathItem)

	spec := &openapi3.T{
		Paths: paths, // paths is already *Paths
	}

	tools := g.GenerateTools(spec)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	// Check first tool (order may vary due to map iteration)
	found := false
	for _, tool := range tools {
		if tool.Name == "getUsers" {
			found = true
			if tool.HTTPMethod != "GET" {
				t.Errorf("expected method 'GET', got '%s'", tool.HTTPMethod)
			}
			if tool.Path != "/users" {
				t.Errorf("expected path '/users', got '%s'", tool.Path)
			}
			if !strings.Contains(tool.Description, "List users") {
				t.Errorf("description should contain 'List users', got '%s'", tool.Description)
			}
		}
	}
	if !found {
		t.Error("expected to find getUsers tool")
	}
}

func TestGenerateToolsWithParameters(t *testing.T) {
	g := NewGenerator()

	paths := openapi3.NewPaths()
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "getUser",
			Summary:     "Get user by ID",
			Parameters: openapi3.Parameters{
				{Value: &openapi3.Parameter{
					Name:        "id",
					In:          "path",
					Required:    true,
					Description: "User ID",
					Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{}},
				}},
			},
		},
	}
	paths.Set("/users/{id}", pathItem)

	spec := &openapi3.T{
		Paths: paths,
	}

	tools := g.GenerateTools(spec)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	tool := tools[0]
	if len(tool.PathParams) != 1 || tool.PathParams[0] != "id" {
		t.Errorf("expected path param 'id', got %v", tool.PathParams)
	}
	if tool.Annotations == nil {
		t.Error("annotations should not be nil")
	}
}

func TestIsDestructiveOperation(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		method      string
		path        string
		operationId string
		destructive bool
	}{
		{"DELETE", "/users/123", "", true},
		{"POST", "/users/123/cancel", "", true},
		{"GET", "/users/123", "", false},
		{"POST", "/users", "CreateUser", false},
		{"DELETE", "/users/123", "DeleteUser", true},
	}

	for _, tc := range tests {
		result := g.isDestructiveOperation(tc.method, tc.path, tc.operationId)
		if result != tc.destructive {
			t.Errorf("isDestructiveOperation(%s, %s, %s) = %v, want %v",
				tc.method, tc.path, tc.operationId, result, tc.destructive)
		}
	}
}

func TestGenerateOperationId(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		path   string
		method string
		want   string
	}{
		{"/users", "get", "GET_users"},
		{"/users/{id}", "get", "GET_users"},
		{"/users/admin", "post", "POST_users_admin"},
	}

	for _, tc := range tests {
		result := g.generateOperationId(tc.path, tc.method)
		if result != tc.want {
			t.Errorf("generateOperationId(%s, %s) = %s, want %s",
				tc.path, tc.method, result, tc.want)
		}
	}
}

func TestToolAnnotations(t *testing.T) {
	g := NewGenerator()

	paths := openapi3.NewPaths()
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "getUsers",
			Summary:     "List users",
		},
		Delete: &openapi3.Operation{
			OperationID: "deleteUsers",
			Summary:     "Delete users",
		},
	}
	paths.Set("/users", pathItem)

	spec := &openapi3.T{
		Paths: paths,
	}

	tools := g.GenerateTools(spec)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	// Find GET and DELETE tools
	var getTool, deleteTool *ToolDefinition
	for _, tool := range tools {
		if tool.Name == "getUsers" {
			getTool = &tool
		}
		if tool.Name == "deleteUsers" {
			deleteTool = &tool
		}
	}

	if getTool == nil || getTool.Annotations == nil {
		t.Fatal("getUsers tool or annotations is nil")
	}

	if !getTool.Annotations.ReadOnlyHint {
		t.Error("GET should be read-only")
	}
	if !getTool.Annotations.IdempotentHint {
		t.Error("GET should be idempotent")
	}

	if deleteTool == nil || deleteTool.Annotations == nil {
		t.Fatal("deleteUsers tool or annotations is nil")
	}

	if deleteTool.Annotations.DestructiveHint == nil || !*deleteTool.Annotations.DestructiveHint {
		t.Error("DELETE should be destructive")
	}
	if !deleteTool.Annotations.IdempotentHint {
		t.Error("DELETE should be idempotent")
	}
}

func TestBuildDescription(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		summary     string
		description string
		wantContain string
	}{
		{"List users", "Returns all users", "List users — Returns all users"},
		{"List users", "", "List users"},
		{"", "Returns all users", "Returns all users"},
		{"", "", ""},
	}

	for _, tc := range tests {
		operation := &openapi3.Operation{
			Summary:     tc.summary,
			Description: tc.description,
		}
		result := g.buildDescription(operation)
		if tc.wantContain != "" && !strings.Contains(result, tc.wantContain) {
			t.Errorf("buildDescription() = %s, should contain %s", result, tc.wantContain)
		}
	}
}
