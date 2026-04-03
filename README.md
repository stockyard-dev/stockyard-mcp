# Stockyard MCP Server

[Model Context Protocol](https://modelcontextprotocol.io) server for managing Stockyard self-hosted tools from AI editors like Claude Desktop, Cursor, Windsurf, and others.

## Install

```bash
curl -fsSL https://stockyard.dev/stockyard-mcp/install.sh | sh
```

Or build from source:

```bash
go install github.com/stockyard-dev/stockyard-mcp/cmd/stockyard-mcp@latest
```

## Configure

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "stockyard": {
      "command": "stockyard-mcp"
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "stockyard": {
      "command": "stockyard-mcp"
    }
  }
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `list_tools` | List installed Stockyard tools and their status |
| `install_tool` | Install a tool by name |
| `tool_status` | Check health of a running tool |
| `query_api` | Make API calls to running tools |
| `available_tools` | Browse all 150 available tools |

## Example Usage

> "Install the bounty bug tracker and check its status"

> "What Stockyard tools do I have installed?"

> "Query the headcount analytics API for today's visitors"

## About

[Stockyard](https://stockyard.dev) is a collection of 150 self-hosted developer tools. Each ships as a single Go binary with embedded SQLite. No Docker, no Postgres, no external dependencies.
