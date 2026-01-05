// Package repos provides repository management for scmd plugins
package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Repository represents a plugin repository
type Repository struct {
	Name        string    `json:"name" yaml:"name"`
	URL         string    `json:"url" yaml:"url"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool      `json:"enabled" yaml:"enabled"`
	LastUpdated time.Time `json:"last_updated,omitempty" yaml:"last_updated,omitempty"`
}

// Manifest is the repo's scmd-repo.yaml file
type Manifest struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	Description string    `yaml:"description"`
	Author      string    `yaml:"author,omitempty"`
	Homepage    string    `yaml:"homepage,omitempty"`
	Commands    []Command `yaml:"commands"`
}

// Command represents a slash command from a repo
type Command struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Usage       string   `yaml:"usage,omitempty"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Category    string   `yaml:"category,omitempty"`
	File        string   `yaml:"file"` // Path to command YAML file in repo
}

// CommandSpec is the full command specification from a YAML file
type CommandSpec struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Usage       string            `yaml:"usage"`
	Aliases     []string          `yaml:"aliases,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Author      string            `yaml:"author,omitempty"`
	Args        []ArgSpec         `yaml:"args,omitempty"`
	Flags       []FlagSpec        `yaml:"flags,omitempty"`
	Prompt      PromptSpec        `yaml:"prompt"`
	Model       ModelSpec         `yaml:"model,omitempty"`
	Examples    []string          `yaml:"examples,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

// ArgSpec defines a command argument
type ArgSpec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default,omitempty"`
}

// FlagSpec defines a command flag
type FlagSpec struct {
	Name        string `yaml:"name"`
	Short       string `yaml:"short,omitempty"`
	Description string `yaml:"description"`
	Default     string `yaml:"default,omitempty"`
}

// PromptSpec defines the prompt template
type PromptSpec struct {
	System   string `yaml:"system,omitempty"`
	Template string `yaml:"template"`
}

// ModelSpec defines model preferences
type ModelSpec struct {
	Preferred   string  `yaml:"preferred,omitempty"`
	MinContext  int     `yaml:"min_context,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
}

// Manager manages repositories
type Manager struct {
	mu         sync.RWMutex
	repos      map[string]*Repository
	dataDir    string
	httpClient *http.Client
}

// NewManager creates a new repository manager
func NewManager(dataDir string) *Manager {
	return &Manager{
		repos:   make(map[string]*Repository),
		dataDir: dataDir,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// reposFile returns the path to repos.json
func (m *Manager) reposFile() string {
	return filepath.Join(m.dataDir, "repos.json")
}

// Load loads repositories from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.reposFile())
	if err != nil {
		if os.IsNotExist(err) {
			// Add default repos
			m.repos["official"] = &Repository{
				Name:        "official",
				URL:         "https://raw.githubusercontent.com/scmd/commands/main",
				Description: "Official scmd commands",
				Enabled:     true,
			}
			return nil
		}
		return err
	}

	var repos []*Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		return err
	}

	for _, r := range repos {
		m.repos[r.Name] = r
	}

	return nil
}

// Save saves repositories to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return err
	}

	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		repos = append(repos, r)
	}

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.reposFile(), data, 0644)
}

// Add adds a new repository
func (m *Manager) Add(name, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repos[name]; exists {
		return fmt.Errorf("repository '%s' already exists", name)
	}

	m.repos[name] = &Repository{
		Name:    name,
		URL:     url,
		Enabled: true,
	}

	return nil
}

// Remove removes a repository
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repos[name]; !exists {
		return fmt.Errorf("repository '%s' not found", name)
	}

	delete(m.repos, name)
	return nil
}

// Get returns a repository by name
func (m *Manager) Get(name string) (*Repository, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.repos[name]
	return r, ok
}

// List returns all repositories
func (m *Manager) List() []*Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		repos = append(repos, r)
	}
	return repos
}

// FetchManifest fetches and parses a repo's manifest
func (m *Manager) FetchManifest(ctx context.Context, repo *Repository) (*Manifest, error) {
	url := repo.URL + "/scmd-repo.yaml"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	return &manifest, nil
}

// FetchCommand fetches a command spec from a repo
func (m *Manager) FetchCommand(ctx context.Context, repo *Repository, cmdPath string) (*CommandSpec, error) {
	url := repo.URL + "/" + cmdPath

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch command: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var spec CommandSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}

	return &spec, nil
}

// SearchCommands searches for commands across all repos
func (m *Manager) SearchCommands(ctx context.Context, query string) ([]SearchResult, error) {
	m.mu.RLock()
	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		if r.Enabled {
			repos = append(repos, r)
		}
	}
	m.mu.RUnlock()

	var results []SearchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(r *Repository) {
			defer wg.Done()

			manifest, err := m.FetchManifest(ctx, r)
			if err != nil {
				return
			}

			for _, cmd := range manifest.Commands {
				if matchesQuery(cmd, query) {
					mu.Lock()
					results = append(results, SearchResult{
						Repo:    r.Name,
						Command: cmd,
					})
					mu.Unlock()
				}
			}
		}(repo)
	}

	wg.Wait()
	return results, nil
}

// SearchResult represents a search result
type SearchResult struct {
	Repo    string
	Command Command
}

func matchesQuery(cmd Command, query string) bool {
	if query == "" {
		return true
	}
	// Simple substring match
	return contains(cmd.Name, query) ||
		contains(cmd.Description, query) ||
		contains(cmd.Category, query)
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// InstallCommand saves a command spec to local storage
func (m *Manager) InstallCommand(spec *CommandSpec, installDir string) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	filename := spec.Name + ".yaml"
	filepath := filepath.Join(installDir, filename)

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("write command: %w", err)
	}

	return nil
}

// LoadInstalledCommands loads all installed commands from local storage
func (m *Manager) LoadInstalledCommands(installDir string) ([]*CommandSpec, error) {
	entries, err := os.ReadDir(installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var commands []*CommandSpec
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 5 || name[len(name)-5:] != ".yaml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(installDir, name))
		if err != nil {
			continue
		}

		var spec CommandSpec
		if err := yaml.Unmarshal(data, &spec); err != nil {
			continue
		}

		commands = append(commands, &spec)
	}

	return commands, nil
}

// UninstallCommand removes an installed command
func (m *Manager) UninstallCommand(name, installDir string) error {
	filename := name + ".yaml"
	filepath := filepath.Join(installDir, filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command '%s' is not installed", name)
		}
		return err
	}

	return nil
}
