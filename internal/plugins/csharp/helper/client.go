package helper

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Client is a high-level typed interface to the roslyn-helper subprocess.
type Client struct {
	proc *Process
}

// NewClient locates the helper binary, starts the process, and verifies it
// responds to ping. Returns an error if any step fails.
func NewClient() (*Client, error) {
	helperPath, err := LocateHelper()
	if err != nil {
		return nil, err
	}

	proc, err := Start(helperPath)
	if err != nil {
		return nil, fmt.Errorf("start roslyn helper: %w", err)
	}

	c := &Client{proc: proc}
	if _, err := c.Ping(); err != nil {
		_ = proc.Close()
		return nil, fmt.Errorf("roslyn helper ping failed: %w", err)
	}
	return c, nil
}

// Close shuts down the helper process.
func (c *Client) Close() error {
	return c.proc.Close()
}

// --- Result types ---

// PingResult is returned by the ping method.
type PingResult struct {
	Pong    bool   `json:"pong"`
	Version string `json:"version"`
}

// LoadSolutionResult is returned by the loadSolution method.
type LoadSolutionResult struct {
	Loaded        bool `json:"loaded"`
	ProjectCount  int  `json:"projectCount"`
	DocumentCount int  `json:"documentCount"`
}

// ProjectSummary is the top-level result of the projectSummary method.
type ProjectSummary struct {
	SolutionPath string        `json:"solutionPath"`
	SolutionName string        `json:"solutionName"`
	Projects     []ProjectInfo `json:"projects"`
}

// ProjectInfo describes a single project in the solution.
type ProjectInfo struct {
	Name              string             `json:"name"`
	Path              string             `json:"path"`
	Type              string             `json:"type"`
	TargetFrameworks  []string           `json:"targetFrameworks"`
	OutputType        string             `json:"outputType"`
	RootNamespace     string             `json:"rootNamespace"`
	DocumentCount     int                `json:"documentCount"`
	ProjectReferences []string           `json:"projectReferences"`
	PackageReferences []PackageReference `json:"packageReferences"`
}

// PackageReference is a NuGet package dependency.
type PackageReference struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// --- Methods ---

// Ping sends a ping to the helper and returns the pong result.
func (c *Client) Ping() (*PingResult, error) {
	raw, err := c.proc.Send("ping", nil)
	if err != nil {
		return nil, wrapRpc("ping", err)
	}
	var r PingResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("ping: decode response: %w", err)
	}
	return &r, nil
}

// LoadSolution instructs the helper to load the given .sln or .csproj file.
func (c *Client) LoadSolution(path string) (*LoadSolutionResult, error) {
	params := map[string]string{"path": path}
	raw, err := c.proc.Send("loadSolution", params)
	if err != nil {
		return nil, wrapRpc("loadSolution", err)
	}
	var r LoadSolutionResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("loadSolution: decode response: %w", err)
	}
	return &r, nil
}

// ProjectSummary retrieves the project summary for the currently loaded solution.
func (c *Client) ProjectSummary() (*ProjectSummary, error) {
	raw, err := c.proc.Send("projectSummary", nil)
	if err != nil {
		return nil, wrapRpc("projectSummary", err)
	}
	var r ProjectSummary
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("projectSummary: decode response: %w", err)
	}
	return &r, nil
}

// --- Outline types ---

// OutlineResult is the structural outline of a single .cs file.
type OutlineResult struct {
	Path           string        `json:"path"`
	Namespace      string        `json:"namespace"`
	LineCount      int           `json:"lineCount"`
	Usings         []string      `json:"usings"`
	Types          []OutlineType `json:"types"`
	HasSyntaxErrors bool         `json:"hasSyntaxErrors"`
}

// OutlineType describes a type (class, interface, struct, record, enum) in the file.
type OutlineType struct {
	Kind      string          `json:"kind"`
	Name      string          `json:"name"`
	Modifiers []string        `json:"modifiers"`
	BaseTypes []string        `json:"baseTypes"`
	Members   []OutlineMember `json:"members"`
	Nested    []OutlineType   `json:"nested"`
}

// OutlineMember describes a member of a type (method, property, field, event, constructor).
type OutlineMember struct {
	Kind       string   `json:"kind"`
	Signature  string   `json:"signature"`
	Modifiers  []string `json:"modifiers"`
	Line       int      `json:"line"`
	IsObsolete bool     `json:"isObsolete,omitempty"`
}

// Outline requests a structural outline of the given .cs file.
// Does not require a solution to be loaded.
func (c *Client) Outline(path string) (*OutlineResult, error) {
	params := map[string]string{"path": path}
	raw, err := c.proc.Send("outline", params)
	if err != nil {
		return nil, wrapRpc("outline", err)
	}
	var r OutlineResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("outline: decode response: %w", err)
	}
	return &r, nil
}

// wrapRpc wraps RpcError values into user-friendly messages.
func wrapRpc(method string, err error) error {
	var rpcErr *RpcError
	if errors.As(err, &rpcErr) {
		switch rpcErr.Code {
		case "E_NOT_FOUND":
			return fmt.Errorf("solution not found: %s", rpcErr.Message)
		case "E_LOAD_FAILED":
			return fmt.Errorf("failed to load solution: %s", rpcErr.Message)
		case "E_INVALID_PARAMS":
			return fmt.Errorf("invalid request: %s", rpcErr.Message)
		default:
			return fmt.Errorf("%s failed: %s", method, rpcErr.Message)
		}
	}
	return err
}
