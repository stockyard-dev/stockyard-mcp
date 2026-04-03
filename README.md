# Stockyard MCP Server

[Model Context Protocol](https://modelcontextprotocol.io/) server for [Stockyard](https://stockyard.dev) tools. Use AI editors like Claude Desktop, Cursor, or Windsurf to interact with your Stockyard tools.

## Install

```bash
npm install -g @stockyard-dev/mcp-server
```

## Configure

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "stockyard": {
      "command": "stockyard-mcp"
    }
  }
}
```

## Tools

| Tool | Description |
|------|-------------|
| `stockyard_list_tools` | List installed tools and their status |
| `stockyard_health` | Check health of a specific tool |
| `stockyard_query` | Query any tool's API |
| `stockyard_create` | Create items (issues, contacts, flags) |
| `stockyard_proxy_stats` | Get LLM proxy statistics |

## Examples

Ask Claude:
- "List my running Stockyard tools"
- "Create a bug in Bounty: login page is broken"
- "Show me today's LLM costs from Trough"
- "Check if Bellwether is healthy"
- "Turn on the dark-mode feature flag in Salt Lick"

## License

Apache 2.0
