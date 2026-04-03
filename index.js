#!/usr/bin/env node
// Stockyard MCP Server — Model Context Protocol server for Stockyard tools
// Allows AI editors (Claude, Cursor, Windsurf) to interact with Stockyard APIs

const http = require('http');
const readline = require('readline');

const STOCKYARD_URL = process.env.STOCKYARD_URL || 'http://localhost:4200';

// Tool definitions
const TOOLS = [
  {
    name: "stockyard_list_tools",
    description: "List all installed Stockyard tools and their status",
    inputSchema: { type: "object", properties: {} }
  },
  {
    name: "stockyard_health",
    description: "Check health of a Stockyard tool",
    inputSchema: {
      type: "object",
      properties: { tool: { type: "string", description: "Tool slug (e.g., bounty, trough)" } },
      required: ["tool"]
    }
  },
  {
    name: "stockyard_query",
    description: "Query a Stockyard tool's API endpoint",
    inputSchema: {
      type: "object",
      properties: {
        tool: { type: "string", description: "Tool slug" },
        endpoint: { type: "string", description: "API endpoint path (e.g., /api/issues)" },
        method: { type: "string", description: "HTTP method", enum: ["GET", "POST", "PUT", "DELETE"], default: "GET" },
        body: { type: "string", description: "JSON request body for POST/PUT" }
      },
      required: ["tool", "endpoint"]
    }
  },
  {
    name: "stockyard_create",
    description: "Create an item in a Stockyard tool (issue, contact, flag, etc.)",
    inputSchema: {
      type: "object",
      properties: {
        tool: { type: "string", description: "Tool slug" },
        data: { type: "object", description: "Item data to create" }
      },
      required: ["tool", "data"]
    }
  },
  {
    name: "stockyard_proxy_stats",
    description: "Get LLM proxy statistics (requests, costs, latency)",
    inputSchema: { type: "object", properties: {} }
  }
];

// Tool ports
const PORTS = {
  bounty: 9320, trough: 9700, saltlick: 8670, bellwether: 8650,
  headcount: 8690, strongbox: 8610, corral: 8760, hub: 9800,
  dossier: 8700, cipher: 8770, notebook: 8750, handbook: 8720,
  billfold: 8710, lasso: 8780, paddock: 8680, sentinel: 8730,
  seismograph: 8600, prospector: 8790, assay: 8590, pipeline: 8800
};

function toolUrl(slug, path) {
  const port = PORTS[slug] || 9700;
  return `http://localhost:${port}${path}`;
}

function fetch(url, opts = {}) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const options = {
      hostname: u.hostname, port: u.port, path: u.pathname + u.search,
      method: opts.method || 'GET',
      headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
      timeout: 5000
    };
    const req = http.request(options, res => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => resolve({ status: res.statusCode, data }));
    });
    req.on('error', reject);
    req.on('timeout', () => { req.destroy(); reject(new Error('timeout')); });
    if (opts.body) req.write(opts.body);
    req.end();
  });
}

async function handleTool(name, args) {
  try {
    switch (name) {
      case 'stockyard_list_tools': {
        const results = [];
        for (const [slug, port] of Object.entries(PORTS)) {
          try {
            const r = await fetch(toolUrl(slug, '/api/health'));
            results.push({ slug, port, status: r.status === 200 ? 'running' : 'unhealthy' });
          } catch { results.push({ slug, port, status: 'stopped' }); }
        }
        return JSON.stringify(results, null, 2);
      }
      case 'stockyard_health': {
        const r = await fetch(toolUrl(args.tool, '/api/health'));
        return r.data;
      }
      case 'stockyard_query': {
        const r = await fetch(toolUrl(args.tool, args.endpoint), {
          method: args.method || 'GET',
          body: args.body
        });
        return r.data;
      }
      case 'stockyard_create': {
        const r = await fetch(toolUrl(args.tool, '/api/items'), {
          method: 'POST',
          body: JSON.stringify(args.data)
        });
        return r.data;
      }
      case 'stockyard_proxy_stats': {
        const r = await fetch(`${STOCKYARD_URL}/api/observe/stats`);
        return r.data;
      }
      default:
        return JSON.stringify({ error: `Unknown tool: ${name}` });
    }
  } catch (e) {
    return JSON.stringify({ error: e.message });
  }
}

// MCP stdio transport
const rl = readline.createInterface({ input: process.stdin });
function send(msg) { process.stdout.write(JSON.stringify(msg) + '\n'); }

rl.on('line', async (line) => {
  try {
    const msg = JSON.parse(line);
    
    if (msg.method === 'initialize') {
      send({ jsonrpc: '2.0', id: msg.id, result: {
        protocolVersion: '2024-11-05',
        capabilities: { tools: {} },
        serverInfo: { name: 'stockyard-mcp', version: '1.0.0' }
      }});
    } else if (msg.method === 'tools/list') {
      send({ jsonrpc: '2.0', id: msg.id, result: { tools: TOOLS } });
    } else if (msg.method === 'tools/call') {
      const result = await handleTool(msg.params.name, msg.params.arguments || {});
      send({ jsonrpc: '2.0', id: msg.id, result: {
        content: [{ type: 'text', text: result }]
      }});
    } else if (msg.method === 'notifications/initialized') {
      // no response needed
    } else {
      send({ jsonrpc: '2.0', id: msg.id, error: { code: -32601, message: 'Method not found' } });
    }
  } catch (e) {
    send({ jsonrpc: '2.0', id: null, error: { code: -32700, message: e.message } });
  }
});

process.stderr.write('Stockyard MCP server running on stdio\n');
