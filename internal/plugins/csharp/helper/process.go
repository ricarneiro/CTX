package helper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync/atomic"
)

// Process manages the lifetime of the roslyn-helper subprocess.
// Calls are sequential — no concurrency within a single Process.
type Process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	nextID atomic.Int64
}

// Start launches the helper subprocess and returns a Process ready to use.
func Start(helperPath string) (*Process, error) {
	cmd := exec.Command(helperPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("helper stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("helper stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("helper stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("helper start: %w", err)
	}

	// Drain stderr in background to prevent blocking.
	go func() { _, _ = io.Copy(io.Discard, stderrPipe) }()

	p := &Process{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
	}
	return p, nil
}

// Send sends a JSON-RPC request and returns the raw result JSON.
// Returns an error if the helper returns an RpcError or dies.
func (p *Process) Send(method string, params interface{}) (json.RawMessage, error) {
	id := int(p.nextID.Add(1))

	var rawParams json.RawMessage
	if params != nil {
		var err error
		rawParams, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
	}

	req := Request{ID: id, Method: method, Params: rawParams}
	line, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	line = append(line, '\n')

	if _, err := p.stdin.Write(line); err != nil {
		return nil, fmt.Errorf("helper write (process may have crashed): %w", err)
	}

	respLine, err := p.stdout.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("helper read (process may have crashed): %w", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(respLine), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.ID != id {
		return nil, fmt.Errorf("response id mismatch: got %d, want %d", resp.ID, id)
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Result, nil
}

// Close sends a shutdown request, closes stdin, and waits for the process to exit.
func (p *Process) Close() error {
	// Best-effort shutdown — ignore errors here.
	_, _ = p.Send("shutdown", nil)
	_ = p.stdin.Close()
	_ = p.cmd.Wait()
	return nil
}
