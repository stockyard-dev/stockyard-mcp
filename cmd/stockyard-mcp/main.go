package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := handle(req)
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
	}
}

func handle(req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "stockyard-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "notifications/initialized":
		return Response{JSONRPC: "2.0", ID: req.ID}

	case "tools/list":
		tools := []Tool{
			{
				Name:        "list_tools",
				Description: "List all installed Stockyard tools and their status (running/stopped)",
				InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
			},
			{
				Name:        "install_tool",
				Description: "Install a Stockyard tool by name (e.g., bounty, headcount, trough)",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"tool": map[string]any{"type": "string", "description": "Tool name (e.g., bounty, headcount, strongbox)"},
					},
					"required": []string{"tool"},
				},
			},
			{
				Name:        "tool_status",
				Description: "Check the health and status of a running Stockyard tool",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"tool": map[string]any{"type": "string", "description": "Tool name"},
						"port": map[string]any{"type": "integer", "description": "Port the tool is running on"},
					},
					"required": []string{"tool"},
				},
			},
			{
				Name:        "query_api",
				Description: "Make an API call to a running Stockyard tool",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"url":    map[string]any{"type": "string", "description": "Full URL (e.g., http://localhost:9320/api/issues)"},
						"method": map[string]any{"type": "string", "enum": []string{"GET", "POST"}, "description": "HTTP method"},
						"body":   map[string]any{"type": "string", "description": "Request body (JSON)"},
					},
					"required": []string{"url"},
				},
			},
			{
				Name:        "available_tools",
				Description: "List all 150 available Stockyard tools with descriptions",
				InputSchema: map[string]any{"type": "object", "properties": map[string]any{
					"category": map[string]any{"type": "string", "description": "Filter by category: developer, operations, creator, finance, personal"},
				}},
			},
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": tools}}

	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		json.Unmarshal(req.Params, &params)
		result := callTool(params.Name, params.Arguments)
		return Response{JSONRPC: "2.0", ID: req.ID, Result: result}

	default:
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   map[string]any{"code": -32601, "message": "method not found: " + req.Method},
		}
	}
}

func callTool(name string, args map[string]any) map[string]any {
	switch name {
	case "list_tools":
		return listTools()
	case "install_tool":
		tool, _ := args["tool"].(string)
		return installTool(tool)
	case "tool_status":
		tool, _ := args["tool"].(string)
		port := 0
		if p, ok := args["port"].(float64); ok {
			port = int(p)
		}
		return toolStatus(tool, port)
	case "query_api":
		url, _ := args["url"].(string)
		method, _ := args["method"].(string)
		if method == "" {
			method = "GET"
		}
		body, _ := args["body"].(string)
		return queryAPI(url, method, body)
	case "available_tools":
		cat, _ := args["category"].(string)
		return availableTools(cat)
	default:
		return textResult("Unknown tool: " + name)
	}
}

func listTools() map[string]any {
	// Find installed stockyard-* binaries
	out, err := exec.Command("sh", "-c", "which stockyard-* 2>/dev/null || find /usr/local/bin /usr/bin $HOME/.local/bin -name 'stockyard-*' -executable 2>/dev/null").Output()
	if err != nil || len(out) == 0 {
		// Check discovery directory
		home, _ := os.UserHomeDir()
		entries, err := os.ReadDir(home + "/.stockyard/discovery")
		if err == nil && len(entries) > 0 {
			var tools []string
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".json") {
					data, _ := os.ReadFile(home + "/.stockyard/discovery/" + e.Name())
					tools = append(tools, string(data))
				}
			}
			return textResult("Discovered tools:\n" + strings.Join(tools, "\n"))
		}
		return textResult("No Stockyard tools found. Install with: curl -fsSL stockyard.dev/bounty/install.sh | sh")
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var results []string
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name != "" {
			slug := strings.TrimPrefix(name, "/usr/local/bin/")
			slug = strings.TrimPrefix(slug, "/usr/bin/")
			results = append(results, slug)
		}
	}
	return textResult("Installed tools:\n" + strings.Join(results, "\n"))
}

func installTool(tool string) map[string]any {
	if tool == "" {
		return textResult("Error: tool name required. Example: bounty, headcount, strongbox")
	}
	cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -fsSL https://stockyard.dev/%s/install.sh | sh", tool))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return textResult("Install failed: " + string(out))
	}
	return textResult("Installed stockyard-" + tool + "\n" + string(out))
}

func toolStatus(tool string, port int) map[string]any {
	if port == 0 {
		// Try common ports
		ports := map[string]int{
			"bounty": 9320, "headcount": 8690, "strongbox": 8610,
			"bellwether": 8650, "saltlick": 8670, "corral": 8760,
			"trough": 9700, "hub": 9800, "paddock": 8680,
		}
		if p, ok := ports[tool]; ok {
			port = p
		} else {
			port = 8600
		}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/api/health", port))
	if err != nil {
		return textResult(fmt.Sprintf("stockyard-%s: not running on port %d", tool, port))
	}
	defer resp.Body.Close()
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	data, _ := json.MarshalIndent(body, "", "  ")
	return textResult(fmt.Sprintf("stockyard-%s (port %d): healthy\n%s", tool, port, string(data)))
}

func queryAPI(url, method, body string) map[string]any {
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	var err error
	if method == "POST" && body != "" {
		resp, err = client.Post(url, "application/json", strings.NewReader(body))
	} else {
		resp, err = client.Get(url)
	}
	if err != nil {
		return textResult("Request failed: " + err.Error())
	}
	defer resp.Body.Close()
	var result any
	json.NewDecoder(resp.Body).Decode(&result)
	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, string(data)))
}

func availableTools(category string) map[string]any {
	// Return curated list
	tools := map[string][]string{
		"developer":  {"bounty (bug tracker)", "assay (API testing)", "pipeline (CI/CD)", "corral (webhook testing)", "strongbox (secrets)", "codex (snippets)", "scaffold (project templates)", "lasso (link shortener)", "fence (API gateway)", "gate (auth proxy)"},
		"operations": {"bellwether (uptime)", "paddock (status page)", "sentinel (alerting)", "headcount (analytics)", "outpost (monitoring)", "handbook (wiki)", "chronicle (logging)", "muster (events)", "roster (HR)", "campfire (standups)"},
		"creator":    {"saltlick (feature flags)", "podium (feedback)", "surveyor (forms)", "post (blog)", "brander (brand assets)", "pasture (social scheduling)", "presskit (press resources)", "pulpit (presentations)"},
		"finance":    {"billfold (invoicing)", "dossier (CRM)", "ledger (accounting)", "prospector (lead gen)", "steward (expenses)", "exchequer (billing)", "dividend (affiliates)"},
		"personal":   {"cipher (passwords)", "notebook (notes)", "almanac (journal)", "trailhead (habits)", "curator (bookmarks)", "feedreader (RSS)", "cellar (collections)"},
	}

	if category != "" {
		if t, ok := tools[category]; ok {
			return textResult(strings.Join(t, "\n"))
		}
		return textResult("Unknown category. Options: developer, operations, creator, finance, personal")
	}

	var all []string
	for cat, t := range tools {
		all = append(all, "\n## "+strings.Title(cat))
		all = append(all, t...)
	}
	return textResult("150 available tools at stockyard.dev/tools/\n" + strings.Join(all, "\n"))
}

func textResult(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
}
