// Package mcp implements a Model Context Protocol server for spine graphs.
// It exposes the spine/api operations as MCP tools over JSON-RPC 2.0 on stdio.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/imran31415/spine/api"
)

// JSON-RPC 2.0 types.

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP-specific types.

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	Tools *struct{} `json:"tools"`
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolHandler func(json.RawMessage) (any, error)

// Server is the MCP server wrapping a spine API Manager.
type Server struct {
	mgr   *api.Manager
	tools map[string]toolHandler
	defs  []ToolDefinition
}

// NewServer creates an MCP server backed by the given Manager.
func NewServer(mgr *api.Manager) *Server {
	s := &Server{
		mgr:   mgr,
		tools: make(map[string]toolHandler),
	}
	s.registerTools()
	return s
}

// Run reads JSON-RPC requests from r and writes responses to w.
// It blocks until r is exhausted or an I/O error occurs.
func (s *Server) Run(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			resp := &Response{
				JSONRPC: "2.0",
				ID:      json.RawMessage("null"),
				Error:   &RPCError{Code: -32700, Message: "parse error"},
			}
			if err := writeResponse(w, resp); err != nil {
				return err
			}
			continue
		}

		resp := s.handle(&req)
		if resp == nil {
			// Notification — no response.
			continue
		}
		if err := writeResponse(w, resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) handle(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: InitializeResult{
				ProtocolVersion: "2024-11-05",
				ServerInfo:      ServerInfo{Name: "spine-mcp", Version: "0.1.0"},
				Capabilities:    ServerCapabilities{Tools: &struct{}{}},
			},
		}

	case "notifications/initialized":
		log.Println("client initialized")
		return nil

	case "tools/list":
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{"tools": s.defs},
		}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32602, Message: "invalid params: " + err.Error()},
			}
		}
		handler, ok := s.tools[params.Name]
		if !ok {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", params.Name)},
			}
		}
		result, err := handler(params.Arguments)
		if err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: ToolCallResult{
					Content: []ContentBlock{{Type: "text", Text: err.Error()}},
					IsError: true,
				},
			}
		}
		text, _ := json.Marshal(result)
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ToolCallResult{
				Content: []ContentBlock{{Type: "text", Text: string(text)}},
			},
		}

	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

func (s *Server) addTool(name, description string, schema any, handler toolHandler) {
	s.defs = append(s.defs, ToolDefinition{
		Name:        name,
		Description: description,
		InputSchema: schema,
	})
	s.tools[name] = handler
}

func writeResponse(w io.Writer, resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}
