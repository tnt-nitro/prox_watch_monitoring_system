package pattern

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Registry manages pattern definitions.
type Registry struct {
	patterns map[string]Pattern
}

// NewRegistry creates a new pattern registry.
func NewRegistry() *Registry {
	return &Registry{
		patterns: make(map[string]Pattern),
	}
}

// LoadPatterns loads patterns from a YAML file (metadata only).
func (r *Registry) LoadPatterns(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("pattern file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read pattern file: %w", err)
	}

	var patterns []Pattern
	if err := yaml.Unmarshal(data, &patterns); err != nil {
		return fmt.Errorf("failed to parse pattern file: %w", err)
	}

	// Validate and register patterns
	for _, pattern := range patterns {
		if err := r.validatePattern(pattern); err != nil {
			return fmt.Errorf("invalid pattern %s: %w", pattern.PatternID, err)
		}

		// Check for duplicates
		if _, exists := r.patterns[pattern.PatternID]; exists {
			return fmt.Errorf("duplicate pattern ID: %s", pattern.PatternID)
		}

		r.patterns[pattern.PatternID] = pattern
	}

	return nil
}

// GetPattern retrieves a pattern by ID.
func (r *Registry) GetPattern(patternID string) (Pattern, bool) {
	pattern, exists := r.patterns[patternID]
	return pattern, exists
}

// GetAllPatterns returns all registered patterns.
func (r *Registry) GetAllPatterns() []Pattern {
	patterns := make([]Pattern, 0, len(r.patterns))
	for _, pattern := range r.patterns {
		patterns = append(patterns, pattern)
	}
	return patterns
}

// validatePattern validates a pattern definition.
func (r *Registry) validatePattern(p Pattern) error {
	if p.PatternID == "" {
		return fmt.Errorf("pattern ID cannot be empty")
	}
	if p.Source == "" {
		return fmt.Errorf("pattern source cannot be empty")
	}
	if p.MatchType < MatchTypeKeyword || p.MatchType > MatchTypeEvent {
		return fmt.Errorf("invalid match type")
	}
	return nil
}

// LoadPatternsFromDir loads all pattern files from a directory.
func LoadPatternsFromDir(dir string) (*Registry, error) {
	registry := NewRegistry()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read pattern directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := registry.LoadPatterns(path); err != nil {
			return nil, fmt.Errorf("failed to load patterns from %s: %w", path, err)
		}
	}

	return registry, nil
}
