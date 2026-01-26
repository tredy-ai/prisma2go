package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the prisma2go configuration.
type Config struct {
	// Required
	Input string `yaml:"input"`

	// Optional
	Output  string `yaml:"output"`
	Package string `yaml:"package"`

	// Struct tags
	JSONTags      bool `yaml:"json_tags"`
	DBTags        bool `yaml:"db_tags"`
	JSONOmitempty bool `yaml:"json_omitempty"`

	// Type options
	UUIDPackage    string `yaml:"uuid_package"`
	DecimalPackage string `yaml:"decimal_package"`

	// Behavior
	NoEnumer bool `yaml:"no_enumer"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Package:       "models",
		JSONTags:      true,
		DBTags:        true,
		JSONOmitempty: true,
		UUIDPackage:   "github.com/google/uuid",
	}
}

// LoadConfig loads configuration from a YAML file.
// If path is empty, it searches for prisma2go.yaml or .prisma2go.yaml in the current directory.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		// Search for default config files
		candidates := []string{"prisma2go.yaml", ".prisma2go.yaml"}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}
	}

	if path == "" {
		// No config file found, return defaults
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Input == "" {
		return errors.New("input file is required (use -i flag or set 'input' in config)")
	}
	return nil
}

// OutputDir returns the directory where output files should be written.
// If output is empty (stdout), returns current directory.
func (c *Config) OutputDir() string {
	if c.Output == "" {
		return "."
	}
	return filepath.Dir(c.Output)
}
