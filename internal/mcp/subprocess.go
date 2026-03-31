package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/forgia-labs/forgia/internal/config"
)

// Compile-time interface check.
var _ ToolProvider = (*MCPSubprocessProvider)(nil)

// MCPSubprocessProvider spawns an MCP server as a long-lived subprocess
// and communicates via JSON-RPC 2.0 over stdio.
type MCPSubprocessProvider struct {
	name            string
	command         string
	args            []string
	lazy            bool
	restartOnCrash  bool
	healthCheckTool string
	exposeTools     map[string]bool
	namespace       string

	mu        sync.RWMutex
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	transport *StdioTransport
	tools     []ToolDefinition
	healthy   bool
	started   bool
	stopping  bool

	startMu sync.Mutex
	nextID  atomic.Int64
	pending sync.Map // int → chan jsonrpcResponse

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{} // closed when subprocess exits
	logger *slog.Logger
	env    []string // additional env vars for subprocess
}

// NewMCPSubprocessProvider creates a provider from config. Does NOT start the subprocess.
func NewMCPSubprocessProvider(cfg config.MCPProviderConfig) *MCPSubprocessProvider {
	expose := make(map[string]bool, len(cfg.ExposeTools))
	for _, t := range cfg.ExposeTools {
		expose[t] = true
	}
	return &MCPSubprocessProvider{
		name:            cfg.Namespace,
		command:         cfg.Command,
		args:            cfg.Args,
		lazy:            cfg.Lazy,
		restartOnCrash:  cfg.RestartOnCrash,
		healthCheckTool: cfg.HealthCheck,
		exposeTools:     expose,
		namespace:       cfg.Namespace,
		logger:          slog.With("provider", cfg.Namespace),
	}
}

// Name returns the provider identifier.
func (p *MCPSubprocessProvider) Name() string {
	return p.name
}

// Tools returns cached tool definitions, filtered by ExposeTools if configured.
func (p *MCPSubprocessProvider) Tools() []ToolDefinition {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.exposeTools) == 0 {
		result := make([]ToolDefinition, len(p.tools))
		copy(result, p.tools)
		return result
	}

	var filtered []ToolDefinition
	for _, t := range p.tools {
		if p.exposeTools[t.Name] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// Call invokes a tool via JSON-RPC tools/call and returns the parsed result.
func (p *MCPSubprocessProvider) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	if err := p.ensureRunning(ctx); err != nil {
		return nil, err
	}

	callParams := map[string]any{
		"name":      tool,
		"arguments": params,
	}

	result, err := p.sendRequest(ctx, "tools/call", callParams)
	if err != nil {
		return nil, err
	}

	var parsed any
	if err := json.Unmarshal(result, &parsed); err != nil {
		return nil, fmt.Errorf("mcp provider %s: parse result: %w", p.name, err)
	}
	return parsed, nil
}

// Start initializes the provider. If lazy, defers subprocess spawn until first Call.
func (p *MCPSubprocessProvider) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	if p.lazy {
		p.logger.Info("lazy mode, deferring subprocess start")
		return nil
	}
	return p.doStart(ctx)
}

// Stop sends SIGTERM, waits 5s, then SIGKILL. Cleans up resources.
func (p *MCPSubprocessProvider) Stop() error {
	p.mu.Lock()
	if p.stopping {
		p.mu.Unlock()
		return nil
	}
	p.stopping = true
	p.healthy = false
	p.mu.Unlock()

	p.logger.Info("stopping")

	if p.cancel != nil {
		p.cancel()
	}

	return p.killProcess()
}

// Healthy returns true if the subprocess is alive and initialized.
func (p *MCPSubprocessProvider) Healthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.healthy
}

// ensureRunning guarantees the subprocess is running. Handles lazy start.
func (p *MCPSubprocessProvider) ensureRunning(ctx context.Context) error {
	p.mu.RLock()
	healthy := p.healthy
	stopping := p.stopping
	started := p.started
	p.mu.RUnlock()

	if stopping {
		return fmt.Errorf("mcp provider %s: stopping", p.name)
	}
	if healthy {
		return nil
	}
	if started {
		return fmt.Errorf("mcp provider %s: not ready", p.name)
	}

	// Lazy first start
	p.startMu.Lock()
	defer p.startMu.Unlock()

	// Double-check under lock
	p.mu.RLock()
	healthy = p.healthy
	p.mu.RUnlock()
	if healthy {
		return nil
	}

	return p.doStart(ctx)
}

// doStart spawns the subprocess, initializes it, and caches tool definitions.
func (p *MCPSubprocessProvider) doStart(ctx context.Context) error {
	if err := p.spawn(); err != nil {
		return fmt.Errorf("mcp provider %s: spawn: %w", p.name, err)
	}
	if err := p.initialize(ctx); err != nil {
		p.killProcess()
		return fmt.Errorf("mcp provider %s: initialize: %w", p.name, err)
	}
	if err := p.fetchTools(ctx); err != nil {
		p.killProcess()
		return fmt.Errorf("mcp provider %s: fetch tools: %w", p.name, err)
	}

	p.mu.Lock()
	p.healthy = true
	p.started = true
	p.mu.Unlock()

	p.logger.Info("started", "tools", len(p.tools))

	if p.restartOnCrash {
		go p.monitorLoop()
	}
	return nil
}

// spawn starts the subprocess and sets up pipes and reader goroutine.
func (p *MCPSubprocessProvider) spawn() error {
	cmd := exec.Command(p.command, p.args...)
	if len(p.env) > 0 {
		cmd.Env = p.env
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	done := make(chan struct{})

	p.mu.Lock()
	p.cmd = cmd
	p.stdin = stdin
	p.transport = NewStdioTransport(stdout, stdin)
	p.done = done
	p.mu.Unlock()

	go p.readLoop()

	go func() {
		cmd.Wait()
		p.mu.Lock()
		if !p.stopping {
			p.healthy = false
		}
		p.mu.Unlock()
		close(done)
	}()

	p.logger.Info("subprocess spawned", "pid", cmd.Process.Pid)
	return nil
}

// initialize performs the MCP initialize handshake.
func (p *MCPSubprocessProvider) initialize(ctx context.Context) error {
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "forgia",
			"version": "0.1.0",
		},
	}

	if _, err := p.sendRequest(ctx, "initialize", params); err != nil {
		return fmt.Errorf("initialize handshake: %w", err)
	}

	if err := p.sendNotification("notifications/initialized", nil); err != nil {
		return fmt.Errorf("initialized notification: %w", err)
	}

	return nil
}

// fetchTools retrieves tool definitions from the subprocess via tools/list.
func (p *MCPSubprocessProvider) fetchTools(ctx context.Context) error {
	result, err := p.sendRequest(ctx, "tools/list", nil)
	if err != nil {
		return fmt.Errorf("tools/list: %w", err)
	}

	var resp struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return fmt.Errorf("parse tools: %w", err)
	}

	tools := make([]ToolDefinition, 0, len(resp.Tools))
	for _, t := range resp.Tools {
		tools = append(tools, ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.InputSchema,
			Namespace:   p.namespace,
		})
	}

	p.mu.Lock()
	p.tools = tools
	p.mu.Unlock()

	return nil
}

// sendRequest sends a JSON-RPC request and waits for the response.
func (p *MCPSubprocessProvider) sendRequest(ctx context.Context, method string, params any) (json.RawMessage, error) {
	id := int(p.nextID.Add(1))
	req := newRequest(id, method, params)

	ch := make(chan jsonrpcResponse, 1)
	p.pending.Store(id, ch)
	defer p.pending.Delete(id)

	if err := p.transport.WriteRequest(req); err != nil {
		return nil, fmt.Errorf("mcp provider %s: send %s: %w", p.name, method, err)
	}

	p.mu.RLock()
	done := p.done
	p.mu.RUnlock()

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("mcp provider %s: %w", p.name, resp.Error)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("mcp provider %s: %s: %w", p.name, method, ctx.Err())
	case <-done:
		return nil, fmt.Errorf("mcp provider %s: subprocess exited during %s", p.name, method)
	}
}

// sendNotification sends a JSON-RPC notification (no response expected).
func (p *MCPSubprocessProvider) sendNotification(method string, params any) error {
	notif := jsonrpcNotification{
		JSONRPC: jsonrpcVersion,
		Method:  method,
		Params:  params,
	}
	return p.transport.WriteNotification(notif)
}

// readLoop reads JSON-RPC responses from stdout and dispatches to pending callers.
func (p *MCPSubprocessProvider) readLoop() {
	for {
		resp, err := p.transport.ReadResponse()
		if err != nil {
			return // EOF or read error — subprocess exited
		}
		if resp.ID == nil {
			continue // server notification, skip
		}
		if ch, ok := p.pending.LoadAndDelete(*resp.ID); ok {
			ch.(chan jsonrpcResponse) <- *resp
		}
	}
}

// monitorLoop watches for subprocess crashes and handles restart with exponential backoff.
func (p *MCPSubprocessProvider) monitorLoop() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		p.mu.RLock()
		done := p.done
		p.mu.RUnlock()

		<-done

		p.mu.RLock()
		stopping := p.stopping
		p.mu.RUnlock()
		if stopping {
			return
		}

		p.logger.Error("subprocess crashed, restarting", "backoff", backoff)

		select {
		case <-time.After(backoff):
		case <-p.ctx.Done():
			return
		}

		if err := p.spawn(); err != nil {
			p.logger.Error("restart spawn failed", "error", err)
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
		initErr := p.initialize(ctx)
		if initErr == nil {
			initErr = p.fetchTools(ctx)
		}
		cancel()

		if initErr != nil {
			p.logger.Error("restart init failed", "error", initErr)
			p.killProcess()
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		p.mu.Lock()
		p.healthy = true
		p.mu.Unlock()

		backoff = time.Second
		p.logger.Info("restarted successfully")
	}
}

// killProcess sends SIGTERM, waits 5s, then SIGKILL. Closes stdin.
func (p *MCPSubprocessProvider) killProcess() error {
	p.mu.RLock()
	cmd := p.cmd
	done := p.done
	p.mu.RUnlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	p.logger.Info("killing subprocess", "pid", cmd.Process.Pid)

	// Close stdin
	p.mu.Lock()
	if p.stdin != nil {
		p.stdin.Close()
		p.stdin = nil
	}
	p.mu.Unlock()

	// Send SIGTERM
	_ = cmd.Process.Signal(syscall.SIGTERM)

	// Wait up to 5s for graceful exit
	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		p.logger.Warn("subprocess did not exit after SIGTERM, sending SIGKILL")
		_ = cmd.Process.Kill()
		<-done
		return nil
	}
}
