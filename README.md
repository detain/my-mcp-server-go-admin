# Admin MCP Proxy Server (Go)

A standalone MCP (Model Context Protocol) proxy server for the MyAdmin admin API. This server acts as an MCP intermediary that:

- Fetches the OpenAPI spec from a remote URL
- Exposes MCP tools generated from the spec
- Proxies tool calls to the actual admin API
- Supports OAuth 2.1 protected resource metadata
- Compiles to a single static native binary

## Requirements

- Go 1.23+

## Installation

### Pre-built Binaries

Download the latest release for your platform from the GitHub releases page.

### Build from Source

```bash
git clone https://github.com/detain/my-mcp-server-go-admin.git
cd my-mcp-server-go-admin
make build
```

### Cross-compilation

```bash
# Linux AMD64
make build-linux-amd64

# Linux ARM64
make build-linux-arm64

# macOS AMD64
make build-darwin-amd64

# macOS ARM64 (Apple Silicon)
make build-darwin-arm64

# Windows
make build-windows

# Build all platforms
make build-all
```

## Configuration

Create a `.env` file or set environment variables:

```bash
cp .env.example .env
# Edit .env with your settings
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAPI_SPEC_URL` | URL to fetch the OpenAPI admin spec from | Required |
| `API_BASE_URL` | Base URL of the admin API to proxy to | Required |
| `SESSION_DIR` | Directory for session storage | `/tmp/mcp_admin_sessions` |
| `CACHE_DIR` | Directory for cached tool definitions | `/tmp/mcp_admin_cache` |
| `SERVER_NAME` | MCP server name | `myadmin-admin-mcp` |
| `SERVER_VERSION` | MCP server version | `1.0.0` |
| `BEARER_TOKEN` | Bearer token for stdio mode auth | - |
| `API_KEY` | API key for stdio mode auth | - |
| `SESSION_ID` | Session ID for stdio mode auth | - |
| `AUTH_SERVER_URL` | OAuth authorization server URL | - |
| `SERVER_URL` | Public URL of this server | - |

## Running

### HTTP Mode (Default)

Runs as an HTTP server when connected to a TTY:

```bash
./bin/mcp-proxy-admin
```

The server will start on port 8080 by default.

### STDIO Mode

Runs as a stdio server when input is piped (for Claude Desktop, Cursor, etc.):

```bash
cat /dev/null | ./bin/mcp-proxy-admin
```

Or with environment variables:

```bash
OPENAPI_SPEC_URL=https://my.interserver.net/admin/spec/openapi-admin.yaml \
API_BASE_URL=https://my.interserver.net/apiv2/admin \
BEARER_TOKEN=your_token \
./bin/mcp-proxy-admin
```

### Detecting Transport Mode

The server automatically detects the transport mode:

- **TTY connected** (keyboard input) → HTTP server on port 8080
- **Pipe/redirect** (stdin from file/command) → STDIO transport

## Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `/mcp` | POST | MCP JSON-RPC endpoint |
| `/mcp` | GET | SSE streaming endpoint |
| `/.well-known/oauth-protected-resource` | GET | OAuth 2.1 protected resource metadata |
| `/health` | GET | Health check |

## Authentication

The proxy supports multiple authentication methods, checked in order:

1. **Bearer Token** - `Authorization: Bearer <token>`
2. **API Key** - `X-API-KEY: <key>`
3. **Session ID** - `sessionid: <session_id>`

For HTTP mode, auth headers come from the incoming request.
For STDIO mode, auth headers come from environment variables.

### Required Headers Added

When proxying to the upstream API:

- `X-API-APP: 1` - Short-circuits rate limiting for MCP callers
- `X-Request-Id` - Request tracing ID

## Claude Desktop Integration (STDIO)

Add to your Claude Desktop configuration:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "myadmin-admin": {
      "command": "/path/to/mcp-proxy-admin",
      "env": {
        "OPENAPI_SPEC_URL": "https://my.interserver.net/admin/spec/openapi-admin.yaml",
        "API_BASE_URL": "https://my.interserver.net/apiv2/admin",
        "BEARER_TOKEN": "your_token_here"
      }
    }
  }
}
```

## Cursor Integration

Add to Cursor settings (Settings → MCP Servers → Add server):

```json
{
  "mcpServers": {
    "myadmin-admin": {
      "command": "/path/to/mcp-proxy-admin",
      "env": {
        "OPENAPI_SPEC_URL": "https://my.interserver.net/admin/spec/openapi-admin.yaml",
        "API_BASE_URL": "https://my.interserver.net/apiv2/admin",
        "BEARER_TOKEN": "your_token_here"
      }
    }
  }
}
```

## Streamable HTTP Mode (Remote Server)

For connecting to a remote MCP server over HTTP:

```json
{
  "mcpServers": {
    "myadmin-admin": {
      "type": "streamable-http",
      "url": "https://mcp.example.com/mcp",
      "headers": {
        "Authorization": "Bearer your_token_here"
      }
    }
  }
}
```

## Web Server Configuration

### Apache

```apache
<VirtualHost *:443>
    ServerName mcp.example.com
    DocumentRoot /var/www/mcp-proxy-admin

    <Location /mcp>
        ProxyPass http://localhost:8080/mcp
        ProxyPassReverse http://localhost:8080/mcp
    </Location>

    <Location /.well-known/oauth-protected-resource>
        ProxyPass http://localhost:8080/.well-known/oauth-protected-resource
        ProxyPassReverse http://localhost:8080/.well-known/oauth-protected-resource
    </Location>
</VirtualHost>
```

### Nginx

```nginx
server {
    listen 443 ssl;
    server_name mcp.example.com;

    location /mcp {
        proxy_pass http://localhost:8080/mcp;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    location /.well-known/oauth-protected-resource {
        proxy_pass http://localhost:8080/.well-known/oauth-protected-resource;
    }
}
```

## Caching

Tool definitions are cached in `CACHE_DIR` for 1 hour to improve startup time. To clear the cache:

```bash
rm -rf /tmp/mcp_admin_cache/*
```

## OAuth 2.1 Compliance

This server implements the OAuth 2.1 protected resource specification:

- Exposes `/.well-known/oauth-protected-resource` metadata endpoint
- Supports Bearer token authentication
- Returns `WWW-Authenticate: Bearer realm="mcp"` header on 401 responses

## Building

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Build with version info
VERSION=1.0.0 make build-version
```

## License

Proprietary - InterServer
