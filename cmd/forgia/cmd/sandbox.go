package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var sandboxRunner string

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Manage sandbox authentication and status",
}

var sandboxLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate a runner inside the sandbox",
	Long: `Stores credentials for use in sandboxed SDD execution.

Supports two modes:
  - Interactive OAuth login (for Max subscription / free tier)
  - API key configuration (for pay-per-token usage)

Credentials are stored in ~/.forgia/sandbox-auth/<runner>/ and mounted
read-only into containers during forgia exec --sandbox.`,
	RunE: runSandboxLogin,
}

var sandboxStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sandbox authentication status for all runners",
	RunE:  runSandboxStatus,
}

var sandboxLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove sandbox credentials for a runner",
	RunE:  runSandboxLogout,
}

func init() {
	sandboxLoginCmd.Flags().StringVar(&sandboxRunner, "runner", "claude", "Runner to authenticate (claude, openhands, codex)")
	sandboxLogoutCmd.Flags().StringVar(&sandboxRunner, "runner", "claude", "Runner to remove credentials for")

	sandboxCmd.AddCommand(sandboxLoginCmd)
	sandboxCmd.AddCommand(sandboxStatusCmd)
	sandboxCmd.AddCommand(sandboxLogoutCmd)
	rootCmd.AddCommand(sandboxCmd)
}

// sandboxAuthDir returns the path to ~/.forgia/sandbox-auth/<runner>/
func sandboxAuthDir(runner string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home dir: %w", err)
	}
	return filepath.Join(home, ".forgia", "sandbox-auth", runner), nil
}

func runSandboxLogin(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	logger := slog.With("command", "sandbox-login", "runner", sandboxRunner)

	authDir, err := sandboxAuthDir(sandboxRunner)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(authDir, 0o700); err != nil {
		return fmt.Errorf("create auth dir: %w", err)
	}

	fmt.Printf("=== Sandbox Login (%s) ===\n\n", sandboxRunner)
	fmt.Println("Choose auth method:")
	fmt.Println("  1. API key (recommended for sandbox)")
	fmt.Println("  2. Interactive OAuth login (Max subscription)")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choice [1/2]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1", "":
		return loginAPIKey(authDir, sandboxRunner, reader)
	case "2":
		return loginOAuth(ctx, authDir, sandboxRunner, logger)
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}

// loginAPIKey prompts for an API key and stores it.
func loginAPIKey(authDir, runner string, reader *bufio.Reader) error {
	var keyEnvName string
	switch runner {
	case "claude":
		keyEnvName = "ANTHROPIC_API_KEY"
	case "openhands":
		fmt.Println("\nOpenHands needs a model provider API key.")
		fmt.Println("Options: ANTHROPIC_API_KEY, OPENAI_API_KEY, or other provider key.")
		fmt.Print("Env var name [ANTHROPIC_API_KEY]: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = "ANTHROPIC_API_KEY"
		}
		keyEnvName = name
	case "codex":
		keyEnvName = "OPENAI_API_KEY"
	default:
		keyEnvName = "API_KEY"
	}

	// Check if already in environment.
	if envVal := os.Getenv(keyEnvName); envVal != "" {
		fmt.Printf("\n%s found in environment. Use it? [Y/n]: ", keyEnvName)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm == "" || confirm == "y" || confirm == "yes" {
			return writeEnvFile(authDir, keyEnvName, envVal)
		}
	}

	fmt.Printf("\nEnter %s: ", keyEnvName)
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("empty key")
	}

	return writeEnvFile(authDir, keyEnvName, key)
}

// writeEnvFile writes a KEY=VALUE .env file to the auth dir.
func writeEnvFile(authDir, key, value string) error {
	envPath := filepath.Join(authDir, ".env")
	content := fmt.Sprintf("%s=%s\n", key, value)
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	fmt.Printf("\n✓ Credentials saved to %s\n", envPath)
	fmt.Println("  Sandbox runs will use this key automatically.")
	return nil
}

// loginOAuth runs an interactive login inside a Docker container.
func loginOAuth(_ interface{}, authDir, runner string, logger *slog.Logger) error {
	if runner != "claude" {
		return fmt.Errorf("OAuth login is only supported for claude runner")
	}

	// Check Docker available.
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found — OAuth login requires Docker")
	}

	// Determine image.
	image := "forgia-sandbox:latest"

	fmt.Println("\n→ Starting Claude Code in container for OAuth login...")
	fmt.Println("  Complete the login flow, then type /exit to finish.")

	// Run interactive container with persistent auth volume.
	// --network=host needed for OAuth callback on localhost.
	c := exec.Command("docker", "run", "-it", "--rm",
		"--network=host",
		"-v", authDir+":/home/forgia/.claude",
		image,
	)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("OAuth login failed: %w", err)
	}

	// Verify credentials were saved.
	credPath := filepath.Join(authDir, ".credentials.json")
	if _, err := os.Stat(credPath); err != nil {
		// Check if there's any auth file.
		entries, _ := os.ReadDir(authDir)
		if len(entries) == 0 {
			return fmt.Errorf("no credentials saved — login may not have completed")
		}
	}

	fmt.Println("\n✓ OAuth credentials saved.")
	fmt.Println("  Sandbox runs will use your Max subscription automatically.")
	logger.Info("OAuth login completed", "auth_dir", authDir)
	return nil
}

func runSandboxStatus(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	baseDir := filepath.Join(home, ".forgia", "sandbox-auth")

	fmt.Println("=== Sandbox Auth ===")

	runners := []struct {
		name    string
		keyName string
	}{
		{"claude", "ANTHROPIC_API_KEY"},
		{"openhands", "ANTHROPIC_API_KEY"},
		{"codex", "OPENAI_API_KEY"},
	}

	for _, r := range runners {
		authDir := filepath.Join(baseDir, r.name)

		// Check .env file (API key).
		envPath := filepath.Join(authDir, ".env")
		if data, err := os.ReadFile(envPath); err == nil {
			content := strings.TrimSpace(string(data))
			if strings.Contains(content, "=") {
				parts := strings.SplitN(content, "=", 2)
				maskedKey := parts[1]
				if len(maskedKey) > 8 {
					maskedKey = maskedKey[:4] + "..." + maskedKey[len(maskedKey)-4:]
				}
				fmt.Printf("  %-12s ✓ API key (%s=%s)\n", r.name+":", parts[0], maskedKey)
				continue
			}
		}

		// Check OAuth credentials.
		credPath := filepath.Join(authDir, ".credentials.json")
		if _, err := os.Stat(credPath); err == nil {
			fmt.Printf("  %-12s ✓ OAuth credentials\n", r.name+":")
			continue
		}

		// Check any files in auth dir.
		entries, err := os.ReadDir(authDir)
		if err == nil && len(entries) > 0 {
			fmt.Printf("  %-12s ✓ credentials present (%d files)\n", r.name+":", len(entries))
			continue
		}

		// Check env var on host.
		if os.Getenv(r.keyName) != "" {
			fmt.Printf("  %-12s ○ using host env %s (not persisted for sandbox)\n", r.name+":", r.keyName)
			continue
		}

		fmt.Printf("  %-12s ✗ not configured — run: forgia sandbox login --runner=%s\n", r.name+":", r.name)
	}

	// Check local runners that don't need auth.
	if _, err := exec.LookPath("ollama"); err == nil {
		fmt.Printf("  %-12s ○ no auth needed (local)\n", "ollama:")
	}

	fmt.Println()
	return nil
}

func runSandboxLogout(cmd *cobra.Command, args []string) error {
	authDir, err := sandboxAuthDir(sandboxRunner)
	if err != nil {
		return err
	}

	if _, err := os.Stat(authDir); os.IsNotExist(err) {
		fmt.Printf("No credentials found for %s.\n", sandboxRunner)
		return nil
	}

	// Confirm.
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Remove sandbox credentials for %s? [y/N]: ", sandboxRunner)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := os.RemoveAll(authDir); err != nil {
		return fmt.Errorf("remove credentials: %w", err)
	}

	fmt.Printf("✓ Credentials removed for %s.\n", sandboxRunner)
	return nil
}
