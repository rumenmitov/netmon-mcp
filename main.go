package main

import (
	"context"
	"log"
	"os"

	//"encoding/json"
	"fmt"

	//"time"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP Server instance
	s := server.NewMCPServer("HTTP-test-mcp_server",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)
	transport := os.Getenv("MCP_TRANSPORT")
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8080"
	}

	calcTool := mcp.NewTool("calculate",
		mcp.WithDescription("Perform arithmetic operations"),
		mcp.WithString("arithmetic_operation",
			mcp.Required(),
			mcp.Enum("add", "subtract", "multiply", "divide"),
			mcp.Description("The arithmetic operation to perform"),
		),
		mcp.WithNumber("x", mcp.Required(), mcp.Description("First number")),
		mcp.WithNumber("y", mcp.Required(), mcp.Description("Second number")),
	)

	s.AddTool(calcTool, handleCalculate)

	// Start MCP Server via HTTP on port 8080 via localhost
	switch transport {
	case "http":
		httpServer := server.NewStreamableHTTPServer(s,
			server.WithEndpointPath("/mcp"))

		log.Println("Listening on port:")
		if err := httpServer.Start(":" + port); err != nil {
			log.Fatal(err)
		}
	case "sse":
		sseServer := server.NewSSEServer(s)
		if err := sseServer.Start(":" + port); err != nil {
			log.Fatal(err)
		}
	default:
		log.Println("Listening on stdio...")
		if err := server.ServeStdio(s); err != nil {
			log.Fatal(err)
		}
	}

}

func handleCalculate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters using helper methods
	operation, err := req.RequireString("arithmetic_operation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	x, err := req.RequireFloat("x")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	y, err := req.RequireFloat("y")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Perform calculation
	var result float64
	switch operation {
	case "add":
		result = x + y
	case "subtract":
		result = x - y
	case "multiply":
		result = x * y
	case "divide":
		if y == 0 {
			return mcp.NewToolResultError("division by zero"), nil
		}
		result = x / y
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown operation: %s", operation)), nil
	}

	// Return result
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}
