package config

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the kakucho configuration
// Keys are target CLI names, values are maps of extension name to description
type Config map[string]map[string]string

// Extension represents an extension command (runtime only)
type Extension struct {
	Name        string
	Description string
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "kakucho", "config.toml"), nil
}

// Load loads the configuration from the config file
func Load() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	return LoadFrom(configPath)
}

// LoadFrom loads the configuration from a specific path
func LoadFrom(path string) (Config, error) {
	config := make(Config)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save saves the configuration to the config file
func (c Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	return c.SaveTo(configPath)
}

// SaveTo saves the configuration to a specific path
func (c Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(c)
}

// GetExtension returns the extension for a target and name
func (c Config) GetExtension(target, name string) *Extension {
	extensions, ok := c[target]
	if !ok {
		return nil
	}
	desc, ok := extensions[name]
	if !ok {
		return nil
	}
	return &Extension{
		Name:        name,
		Description: desc,
	}
}

// GetExtensions returns all extensions for a target
func (c Config) GetExtensions(target string) []Extension {
	extensions, ok := c[target]
	if !ok {
		return nil
	}
	result := make([]Extension, 0, len(extensions))
	for name, desc := range extensions {
		result = append(result, Extension{
			Name:        name,
			Description: desc,
		})
	}
	return result
}

// GetTargetNames returns all target names
func (c Config) GetTargetNames() []string {
	names := make([]string, 0, len(c))
	for name := range c {
		names = append(names, name)
	}
	return names
}

// AddExtension adds a new extension to a target
func (c Config) AddExtension(target, name, description string) {
	if c[target] == nil {
		c[target] = make(map[string]string)
	}
	c[target][name] = description
}

// RemoveExtension removes an extension from a target
// If the target has no more extensions, the target is also removed
func (c Config) RemoveExtension(target, name string) bool {
	extensions, ok := c[target]
	if !ok {
		return false
	}
	if _, ok := extensions[name]; !ok {
		return false
	}
	delete(extensions, name)
	// Remove target if no extensions left
	if len(extensions) == 0 {
		delete(c, target)
	}
	return true
}

// GetExtensionPath returns the executable path for an extension
// The path is {target}-{name} and is looked up in PATH
func GetExtensionPath(target, name string) (string, error) {
	execName := target + "-" + name
	return exec.LookPath(execName)
}
