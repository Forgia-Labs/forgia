package process

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ShutdownManager coordinates graceful shutdown of all processes.
type ShutdownManager struct {
	processes []*Process
	cleanups  []func()
	mu        sync.Mutex
}

// NewShutdownManager creates a shutdown manager.
func NewShutdownManager() *ShutdownManager {
	return &ShutdownManager{}
}

// Register adds a process to be killed on shutdown.
func (sm *ShutdownManager) Register(p *Process) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.processes = append(sm.processes, p)
}

// OnCleanup adds a cleanup function to run on shutdown.
func (sm *ShutdownManager) OnCleanup(fn func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.cleanups = append(sm.cleanups, fn)
}

// ListenAndShutdown blocks until SIGINT or SIGTERM, then kills all processes.
func (sm *ShutdownManager) ListenAndShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, p := range sm.processes {
		_ = p.Kill()
	}

	for _, fn := range sm.cleanups {
		fn()
	}
}
