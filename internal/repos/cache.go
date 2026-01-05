package repos

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache provides local caching for repos and commands
type Cache struct {
	mu       sync.RWMutex
	dataDir  string
	manifest CacheManifest
}

// CacheManifest tracks cached items
type CacheManifest struct {
	Version    string                 `json:"version"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Repos      map[string]CachedRepo  `json:"repos"`
	Commands   map[string]CachedCmd   `json:"commands"`
	Manifests  map[string]CachedItem  `json:"manifests"`
}

// CachedRepo tracks a cached repository
type CachedRepo struct {
	URL          string    `json:"url"`
	ETag         string    `json:"etag"`
	LastModified string    `json:"last_modified"`
	CachedAt     time.Time `json:"cached_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// CachedCmd tracks a cached command
type CachedCmd struct {
	Repo         string    `json:"repo"`
	Version      string    `json:"version"`
	Hash         string    `json:"hash"`
	CachedAt     time.Time `json:"cached_at"`
	InstalledAt  time.Time `json:"installed_at,omitempty"`
	UpdateAvail  bool      `json:"update_available,omitempty"`
	LatestVer    string    `json:"latest_version,omitempty"`
}

// CachedItem is a generic cached item
type CachedItem struct {
	Hash      string    `json:"hash"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewCache creates a new cache
func NewCache(dataDir string) *Cache {
	return &Cache{
		dataDir: dataDir,
		manifest: CacheManifest{
			Version:   "1.0",
			Repos:     make(map[string]CachedRepo),
			Commands:  make(map[string]CachedCmd),
			Manifests: make(map[string]CachedItem),
		},
	}
}

// Load loads the cache manifest from disk
func (c *Cache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.dataDir, "cache", "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &c.manifest)
}

// Save saves the cache manifest to disk
func (c *Cache) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheDir := filepath.Join(c.dataDir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	c.manifest.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(c.manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, "manifest.json"), data, 0644)
}

// GetManifest retrieves a cached manifest
func (c *Cache) GetManifest(repoURL string) (*Manifest, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := hashURL(repoURL)
	cached, ok := c.manifest.Manifests[key]
	if !ok || time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	path := filepath.Join(c.dataDir, "cache", "manifests", key+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, false
	}

	return &manifest, true
}

// SetManifest caches a manifest
func (c *Cache) SetManifest(repoURL string, manifest *Manifest, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := hashURL(repoURL)
	cacheDir := filepath.Join(c.dataDir, "cache", "manifests")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	path := filepath.Join(cacheDir, key+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	c.manifest.Manifests[key] = CachedItem{
		Hash:      hashData(data),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	return nil
}

// GetCommand retrieves a cached command spec
func (c *Cache) GetCommand(repo, name string) (*CommandSpec, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := repo + "/" + name
	cached, ok := c.manifest.Commands[key]
	if !ok {
		return nil, false
	}

	path := filepath.Join(c.dataDir, "cache", "commands", cached.Hash+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var spec CommandSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, false
	}

	return &spec, true
}

// SetCommand caches a command spec
func (c *Cache) SetCommand(repo, name string, spec *CommandSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheDir := filepath.Join(c.dataDir, "cache", "commands")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	hash := hashData(data)
	path := filepath.Join(cacheDir, hash+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	key := repo + "/" + name
	c.manifest.Commands[key] = CachedCmd{
		Repo:     repo,
		Version:  spec.Version,
		Hash:     hash,
		CachedAt: time.Now(),
	}

	return nil
}

// MarkInstalled marks a command as installed
func (c *Cache) MarkInstalled(repo, name, version string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := repo + "/" + name
	if cached, ok := c.manifest.Commands[key]; ok {
		cached.InstalledAt = time.Now()
		cached.Version = version
		c.manifest.Commands[key] = cached
	} else {
		c.manifest.Commands[key] = CachedCmd{
			Repo:        repo,
			Version:     version,
			InstalledAt: time.Now(),
		}
	}
}

// GetInstalled returns all installed commands
func (c *Cache) GetInstalled() []CachedCmd {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var installed []CachedCmd
	for _, cmd := range c.manifest.Commands {
		if !cmd.InstalledAt.IsZero() {
			installed = append(installed, cmd)
		}
	}
	return installed
}

// CheckUpdates checks for updates to installed commands
func (c *Cache) CheckUpdates(getLatestVersion func(repo, name string) (string, error)) ([]UpdateInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var updates []UpdateInfo
	for key, cmd := range c.manifest.Commands {
		if cmd.InstalledAt.IsZero() {
			continue
		}

		// Parse key to get repo/name
		var repo, name string
		for i := len(key) - 1; i >= 0; i-- {
			if key[i] == '/' {
				repo = key[:i]
				name = key[i+1:]
				break
			}
		}

		latest, err := getLatestVersion(repo, name)
		if err != nil {
			continue
		}

		if latest != cmd.Version {
			cmd.UpdateAvail = true
			cmd.LatestVer = latest
			c.manifest.Commands[key] = cmd

			updates = append(updates, UpdateInfo{
				Repo:       repo,
				Command:    name,
				Current:    cmd.Version,
				Latest:     latest,
				InstalledAt: cmd.InstalledAt,
			})
		}
	}

	return updates, nil
}

// UpdateInfo describes an available update
type UpdateInfo struct {
	Repo        string
	Command     string
	Current     string
	Latest      string
	InstalledAt time.Time
}

// Clear clears the cache
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cacheDir := filepath.Join(c.dataDir, "cache")
	if err := os.RemoveAll(cacheDir); err != nil {
		return err
	}

	c.manifest = CacheManifest{
		Version:   "1.0",
		Repos:     make(map[string]CachedRepo),
		Commands:  make(map[string]CachedCmd),
		Manifests: make(map[string]CachedItem),
	}

	return nil
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var installedCount int
	for _, cmd := range c.manifest.Commands {
		if !cmd.InstalledAt.IsZero() {
			installedCount++
		}
	}

	return CacheStats{
		CachedRepos:    len(c.manifest.Repos),
		CachedCommands: len(c.manifest.Commands),
		CachedManifests: len(c.manifest.Manifests),
		InstalledCommands: installedCount,
		LastUpdated:    c.manifest.UpdatedAt,
	}
}

// CacheStats provides cache statistics
type CacheStats struct {
	CachedRepos       int
	CachedCommands    int
	CachedManifests   int
	InstalledCommands int
	LastUpdated       time.Time
}

func hashURL(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:8])
}

func hashData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

// Lockfile represents a lockfile for reproducible installations
type Lockfile struct {
	Version   string         `json:"version"`
	Generated time.Time      `json:"generated"`
	Commands  []LockedCmd    `json:"commands"`
}

// LockedCmd is a locked command version
type LockedCmd struct {
	Name    string `json:"name"`
	Repo    string `json:"repo"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	URL     string `json:"url"`
}

// GenerateLockfile creates a lockfile from installed commands
func (c *Cache) GenerateLockfile() *Lockfile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lf := &Lockfile{
		Version:   "1.0",
		Generated: time.Now(),
	}

	for key, cmd := range c.manifest.Commands {
		if cmd.InstalledAt.IsZero() {
			continue
		}

		// Parse key
		var repo, name string
		for i := len(key) - 1; i >= 0; i-- {
			if key[i] == '/' {
				repo = key[:i]
				name = key[i+1:]
				break
			}
		}

		lf.Commands = append(lf.Commands, LockedCmd{
			Name:    name,
			Repo:    repo,
			Version: cmd.Version,
			Hash:    cmd.Hash,
		})
	}

	return lf
}

// SaveLockfile saves a lockfile to disk
func SaveLockfile(lf *Lockfile, path string) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadLockfile loads a lockfile from disk
func LoadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}

	return &lf, nil
}

// InstallFromLockfile installs commands from a lockfile
func (m *Manager) InstallFromLockfile(lf *Lockfile, installDir string) error {
	for _, cmd := range lf.Commands {
		repo, ok := m.Get(cmd.Repo)
		if !ok {
			// Try to add the repo
			if cmd.URL != "" {
				if err := m.Add(cmd.Repo, cmd.URL); err != nil {
					return fmt.Errorf("add repo %s: %w", cmd.Repo, err)
				}
				repo, _ = m.Get(cmd.Repo)
			} else {
				return fmt.Errorf("repo %s not found", cmd.Repo)
			}
		}

		// Find the command file
		manifest, err := m.FetchManifest(nil, repo)
		if err != nil {
			return fmt.Errorf("fetch manifest for %s: %w", cmd.Repo, err)
		}

		var cmdFile string
		for _, c := range manifest.Commands {
			if c.Name == cmd.Name {
				cmdFile = c.File
				break
			}
		}

		if cmdFile == "" {
			return fmt.Errorf("command %s not found in %s", cmd.Name, cmd.Repo)
		}

		spec, err := m.FetchCommand(nil, repo, cmdFile)
		if err != nil {
			return fmt.Errorf("fetch command %s: %w", cmd.Name, err)
		}

		// Verify version matches if specified
		if cmd.Version != "" && spec.Version != cmd.Version {
			return fmt.Errorf("version mismatch for %s: wanted %s, got %s",
				cmd.Name, cmd.Version, spec.Version)
		}

		if err := m.InstallCommand(spec, installDir); err != nil {
			return fmt.Errorf("install %s: %w", cmd.Name, err)
		}
	}

	return nil
}
