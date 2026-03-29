package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// ServiceDef describes a service an SDD needs during execution.
type ServiceDef struct {
	Name  string            `yaml:"name"`
	Image string            `yaml:"image"`
	Ports []int             `yaml:"ports,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
}

// ServiceManager starts and stops services declared in SDD boundaries.
type ServiceManager struct {
	services  []ServiceDef
	running   []string // container IDs of running services
	provider  string   // "docker" or "apple-container"
	logger    *slog.Logger
}

// NewServiceManager creates a manager for the given service definitions.
func NewServiceManager(services []ServiceDef, provider string) *ServiceManager {
	return &ServiceManager{
		services: services,
		provider: provider,
		logger:   slog.With("component", "service-manager"),
	}
}

// StartAll starts all declared services and waits for readiness.
// Returns a cleanup function that stops all services.
func (m *ServiceManager) StartAll(ctx context.Context) (cleanup func(), err error) {
	if len(m.services) == 0 {
		return func() {}, nil
	}

	m.logger.InfoContext(ctx, "starting services", "count", len(m.services))

	for _, svc := range m.services {
		containerID, err := m.startService(ctx, svc)
		if err != nil {
			// Stop any already-started services.
			m.StopAll(ctx)
			return nil, fmt.Errorf("start service %s: %w", svc.Name, err)
		}
		m.running = append(m.running, containerID)
		m.logger.InfoContext(ctx, "service started", "name", svc.Name, "container", containerID[:12])
	}

	cleanup = func() {
		m.StopAll(context.Background())
	}
	return cleanup, nil
}

// StopAll stops all running services.
func (m *ServiceManager) StopAll(ctx context.Context) {
	for _, id := range m.running {
		m.stopContainer(ctx, id)
	}
	m.running = nil
}

// startService starts a single service container.
func (m *ServiceManager) startService(ctx context.Context, svc ServiceDef) (string, error) {
	switch m.provider {
	case "docker":
		return m.startDocker(ctx, svc)
	case "apple-container":
		return m.startApple(ctx, svc)
	default:
		return m.startDocker(ctx, svc) // default to docker for services
	}
}

func (m *ServiceManager) startDocker(ctx context.Context, svc ServiceDef) (string, error) {
	args := []string{"run", "-d", "--rm", "--name", "forgia-svc-" + svc.Name}

	for _, port := range svc.Ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", port, port))
	}
	for k, v := range svc.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, svc.Image)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker run: %s: %w", strings.TrimSpace(string(output)), err)
	}

	containerID := strings.TrimSpace(string(output))

	// Wait for readiness (simple health check — container is running).
	if err := m.waitReady(ctx, containerID); err != nil {
		m.stopContainer(ctx, containerID)
		return "", fmt.Errorf("service %s not ready: %w", svc.Name, err)
	}

	return containerID, nil
}

func (m *ServiceManager) startApple(ctx context.Context, svc ServiceDef) (string, error) {
	// Apple Container doesn't support background daemon mode like Docker.
	// For services, fall back to Docker which handles long-running processes better.
	m.logger.InfoContext(ctx, "using docker for service (apple-container doesn't support daemon mode)", "service", svc.Name)
	return m.startDocker(ctx, svc)
}

func (m *ServiceManager) stopContainer(ctx context.Context, id string) {
	cmd := exec.CommandContext(ctx, "docker", "stop", id)
	if err := cmd.Run(); err != nil {
		m.logger.WarnContext(ctx, "failed to stop service container", "id", id[:12], "error", err)
	} else {
		m.logger.InfoContext(ctx, "service stopped", "id", id[:12])
	}
}

// waitReady polls the container until it's running or timeout.
func (m *ServiceManager) waitReady(ctx context.Context, containerID string) error {
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for container %s", containerID[:12])
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", containerID)
			output, err := cmd.CombinedOutput()
			if err == nil && strings.TrimSpace(string(output)) == "true" {
				return nil
			}
		}
	}
}
