package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

// DefaultRegistryURL is the central scmd command registry
const DefaultRegistryURL = "https://registry.scmd.dev/api/v1"

// Registry provides access to the central scmd command registry
type Registry struct {
	URL        string
	httpClient *http.Client
	cache      *RegistryCache
	mu         sync.RWMutex
}

// RegistryCache caches registry data locally
type RegistryCache struct {
	Repos       []RepoEntry    `json:"repos"`
	Commands    []CommandEntry `json:"commands"`
	LastUpdated time.Time      `json:"last_updated"`
	TTL         time.Duration  `json:"ttl"`
}

// RepoEntry represents a repository in the registry
type RepoEntry struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Verified    bool     `json:"verified"`
	Official    bool     `json:"official"`
	Stars       int      `json:"stars"`
	Downloads   int      `json:"downloads"`
	Categories  []string `json:"categories"`
	Tags        []string `json:"tags"`
	UpdatedAt   string   `json:"updated_at"`
}

// CommandEntry represents a command in the registry
type CommandEntry struct {
	Name        string   `json:"name"`
	Repo        string   `json:"repo"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	RatingCount int      `json:"rating_count"`
	Verified    bool     `json:"verified"`
	Featured    bool     `json:"featured"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// SearchOptions for filtering registry searches
type SearchOptions struct {
	Query      string
	Category   string
	Tags       []string
	Verified   bool
	Featured   bool
	SortBy     string // "downloads", "rating", "updated", "name"
	Limit      int
	Offset     int
}

// NewRegistry creates a new registry client
func NewRegistry(url string) *Registry {
	if url == "" {
		url = DefaultRegistryURL
	}
	return &Registry{
		URL: url,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: &RegistryCache{
			TTL: 1 * time.Hour,
		},
	}
}

// SearchRepos searches for repositories in the registry
func (r *Registry) SearchRepos(ctx context.Context, opts SearchOptions) ([]RepoEntry, error) {
	// Build query URL
	url := fmt.Sprintf("%s/repos?q=%s", r.URL, opts.Query)
	if opts.Category != "" {
		url += "&category=" + opts.Category
	}
	if opts.Verified {
		url += "&verified=true"
	}
	if opts.SortBy != "" {
		url += "&sort=" + opts.SortBy
	}
	if opts.Limit > 0 {
		url += fmt.Sprintf("&limit=%d", opts.Limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "scmd/1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		// Fall back to cache if network fails
		return r.getCachedRepos(opts), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return r.getCachedRepos(opts), nil
	}

	var repos []RepoEntry
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	// Update cache
	r.mu.Lock()
	r.cache.Repos = repos
	r.cache.LastUpdated = time.Now()
	r.mu.Unlock()

	return repos, nil
}

// SearchCommands searches for commands in the registry
func (r *Registry) SearchCommands(ctx context.Context, opts SearchOptions) ([]CommandEntry, error) {
	url := fmt.Sprintf("%s/commands?q=%s", r.URL, opts.Query)
	if opts.Category != "" {
		url += "&category=" + opts.Category
	}
	for _, tag := range opts.Tags {
		url += "&tag=" + tag
	}
	if opts.Verified {
		url += "&verified=true"
	}
	if opts.Featured {
		url += "&featured=true"
	}
	if opts.SortBy != "" {
		url += "&sort=" + opts.SortBy
	}
	if opts.Limit > 0 {
		url += fmt.Sprintf("&limit=%d", opts.Limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "scmd/1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return r.getCachedCommands(opts), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return r.getCachedCommands(opts), nil
	}

	var commands []CommandEntry
	if err := json.NewDecoder(resp.Body).Decode(&commands); err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cache.Commands = commands
	r.cache.LastUpdated = time.Now()
	r.mu.Unlock()

	return commands, nil
}

// GetFeatured returns featured/trending commands
func (r *Registry) GetFeatured(ctx context.Context) ([]CommandEntry, error) {
	return r.SearchCommands(ctx, SearchOptions{
		Featured: true,
		SortBy:   "downloads",
		Limit:    10,
	})
}

// GetCategories returns available command categories
func (r *Registry) GetCategories(ctx context.Context) ([]CategoryInfo, error) {
	url := fmt.Sprintf("%s/categories", r.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return defaultCategories(), nil
	}
	req.Header.Set("User-Agent", "scmd/1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return defaultCategories(), nil
	}
	defer resp.Body.Close()

	var categories []CategoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return defaultCategories(), nil
	}

	return categories, nil
}

// CategoryInfo describes a command category
type CategoryInfo struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Count       int    `json:"count"`
}

func defaultCategories() []CategoryInfo {
	return []CategoryInfo{
		{Name: "Git", Slug: "git", Description: "Git workflow commands", Icon: "🔀", Count: 0},
		{Name: "Code", Slug: "code", Description: "Code analysis and generation", Icon: "💻", Count: 0},
		{Name: "DevOps", Slug: "devops", Description: "DevOps and infrastructure", Icon: "🚀", Count: 0},
		{Name: "Data", Slug: "data", Description: "Data processing and analysis", Icon: "📊", Count: 0},
		{Name: "Docs", Slug: "docs", Description: "Documentation generation", Icon: "📝", Count: 0},
		{Name: "Debug", Slug: "debug", Description: "Debugging and troubleshooting", Icon: "🐛", Count: 0},
		{Name: "Text", Slug: "text", Description: "Text processing", Icon: "📄", Count: 0},
		{Name: "Shell", Slug: "shell", Description: "Shell and terminal utilities", Icon: "🐚", Count: 0},
	}
}

// DiscoverFromURL attempts to discover a repo from a URL using .well-known
func (r *Registry) DiscoverFromURL(ctx context.Context, baseURL string) (*Manifest, error) {
	// Try .well-known/scmd.json first (like MCP discovery)
	wellKnownURL := baseURL + "/.well-known/scmd.json"

	req, err := http.NewRequestWithContext(ctx, "GET", wellKnownURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "scmd/1.0")

	resp, err := r.httpClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var manifest Manifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err == nil {
			return &manifest, nil
		}
	}

	// Fall back to scmd-repo.yaml
	return nil, fmt.Errorf("no scmd manifest found at %s", baseURL)
}

// ResolveShorthand resolves shorthand like "official/git-commit" to full repo/command
func (r *Registry) ResolveShorthand(ctx context.Context, shorthand string) (repo, command string, err error) {
	// Parse shorthand format: [registry/]repo/command
	parts := splitPath(shorthand)

	switch len(parts) {
	case 1:
		// Just command name - search registry
		commands, err := r.SearchCommands(ctx, SearchOptions{
			Query: parts[0],
			Limit: 1,
		})
		if err != nil || len(commands) == 0 {
			return "", "", fmt.Errorf("command '%s' not found in registry", parts[0])
		}
		return commands[0].Repo, commands[0].Name, nil

	case 2:
		// repo/command format
		return parts[0], parts[1], nil

	case 3:
		// registry/repo/command format (for future multi-registry support)
		return parts[1], parts[2], nil

	default:
		return "", "", fmt.Errorf("invalid shorthand format: %s", shorthand)
	}
}

func splitPath(s string) []string {
	var parts []string
	var current string
	for _, c := range s {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// getCachedRepos returns cached repos matching the search
func (r *Registry) getCachedRepos(opts SearchOptions) []RepoEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache.Repos == nil {
		return nil
	}

	var results []RepoEntry
	query := toLower(opts.Query)

	for _, repo := range r.cache.Repos {
		if query == "" || contains(repo.Name, query) || contains(repo.Description, query) {
			if opts.Category == "" || containsAny(repo.Categories, opts.Category) {
				if !opts.Verified || repo.Verified {
					results = append(results, repo)
				}
			}
		}
	}

	return results
}

// getCachedCommands returns cached commands matching the search
func (r *Registry) getCachedCommands(opts SearchOptions) []CommandEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache.Commands == nil {
		return nil
	}

	var results []CommandEntry
	query := toLower(opts.Query)

	for _, cmd := range r.cache.Commands {
		if query == "" || contains(cmd.Name, query) || contains(cmd.Description, query) {
			if opts.Category == "" || cmd.Category == opts.Category {
				if !opts.Verified || cmd.Verified {
					if !opts.Featured || cmd.Featured {
						results = append(results, cmd)
					}
				}
			}
		}
	}

	// Sort results
	switch opts.SortBy {
	case "downloads":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Downloads > results[j].Downloads
		})
	case "rating":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Rating > results[j].Rating
		})
	case "name":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results
}

func containsAny(slice []string, s string) bool {
	for _, item := range slice {
		if toLower(item) == toLower(s) {
			return true
		}
	}
	return false
}

// PublishCommand publishes a command to the registry (requires auth)
func (r *Registry) PublishCommand(ctx context.Context, spec *CommandSpec, token string) error {
	url := fmt.Sprintf("%s/commands", r.URL)

	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(
		&bytesReader{data: data, pos: 0},
	))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "scmd/1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

type bytesReader struct {
	data []byte
	pos  int
}

func (b *bytesReader) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}
