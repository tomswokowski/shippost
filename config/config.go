package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const (
	configDirPerm  = 0700
	configFilePerm = 0600
)

// Config holds the X API credentials
type Config struct {
	APIKey       string `json:"api_key"`
	APISecret    string `json:"api_secret"`
	AccessToken  string `json:"access_token"`
	AccessSecret string `json:"access_secret"`
}

// configPath returns the path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "shippost", "config.json"), nil
}

// Exists checks if a valid config file exists
func Exists() bool {
	path, err := configPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Load reads the config from disk
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	// Check if config exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config not found - run 'shippost --setup' to configure")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat config: %w", err)
	}

	// Validate file permissions (warn if too permissive)
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		fmt.Fprintf(os.Stderr, "Warning: config file has overly permissive permissions (%o). Run 'chmod 600 %s'\n", mode, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if !cfg.IsValid() {
		return nil, fmt.Errorf("config is incomplete - run 'shippost --setup' to configure")
	}

	return &cfg, nil
}

// Save writes the config to disk with secure permissions
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create config directory with secure permissions
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, configDirPerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Ensure directory has correct permissions
	if err := os.Chmod(dir, configDirPerm); err != nil {
		return fmt.Errorf("failed to set directory permissions: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with secure permissions
	if err := os.WriteFile(path, data, configFilePerm); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// IsValid checks if all required fields are present
func (c *Config) IsValid() bool {
	return c.APIKey != "" && c.APISecret != "" && c.AccessToken != "" && c.AccessSecret != ""
}

// Cleanup removes the config file and directory
func Cleanup() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Check if config exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("No config found - nothing to clean up")
		return nil
	}

	// Remove config file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove config: %w", err)
	}

	// Try to remove config directory (will fail if not empty, which is fine)
	dir := filepath.Dir(path)
	os.Remove(dir) // ignore error - dir might have other files

	fmt.Println("Credentials removed successfully")
	return nil
}

// RunSetup interactively prompts for credentials
func RunSetup() error {
	fmt.Println("shippost setup")
	fmt.Println("==============")
	fmt.Println("\nYou'll need X API credentials from developer.x.com")
	fmt.Println("Create a project and app with Read+Write permissions.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	cfg := &Config{}

	// API Key
	fmt.Print("API Key (Consumer Key): ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	cfg.APIKey = strings.TrimSpace(apiKey)

	// API Secret (hidden input)
	fmt.Print("API Secret (Consumer Secret): ")
	apiSecretBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read secret: %w", err)
	}
	fmt.Println()
	cfg.APISecret = string(apiSecretBytes)

	// Access Token
	fmt.Print("Access Token: ")
	accessToken, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	cfg.AccessToken = strings.TrimSpace(accessToken)

	// Access Secret (hidden input)
	fmt.Print("Access Token Secret: ")
	accessSecretBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read secret: %w", err)
	}
	fmt.Println()
	cfg.AccessSecret = string(accessSecretBytes)

	if !cfg.IsValid() {
		return fmt.Errorf("all fields are required")
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	path, _ := configPath()
	fmt.Printf("\nConfig saved to %s\n", path)
	fmt.Println("You're ready to post!")

	return nil
}
