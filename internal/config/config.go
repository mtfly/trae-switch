package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Provider struct {
	Name       string   `json:"name"`
	OpenAIBase string   `json:"openai_base"`
	Models     []string `json:"models"`
}

type Config struct {
	Providers      []Provider `json:"providers"`
	ActiveProvider int        `json:"active_provider"`
	TrayMode       bool       `json:"tray_mode"`
	mu             sync.RWMutex
}

var (
	cfg     *Config
	cfgPath string
	once    sync.Once
)

func GetConfigPath() string {
	once.Do(func() {
		exePath, err := os.Executable()
		if err != nil {
			exePath = os.Args[0]
		}
		dir := filepath.Dir(exePath)
		cfgPath = filepath.Join(dir, "config.json")
	})
	return cfgPath
}

func Load() (*Config, error) {
	cfg = &Config{
		Providers:      []Provider{},
		ActiveProvider: 0,
	}

	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func GetProviders() []Provider {
	if cfg == nil {
		return []Provider{}
	}
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	return cfg.Providers
}

func GetActiveProvider() *Provider {
	if cfg == nil {
		return nil
	}
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	if cfg.ActiveProvider >= 0 && cfg.ActiveProvider < len(cfg.Providers) {
		return &cfg.Providers[cfg.ActiveProvider]
	}
	return nil
}

func SetActiveProvider(index int) error {
	if cfg == nil {
		return nil
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if index >= 0 && index < len(cfg.Providers) {
		cfg.ActiveProvider = index
		return cfg.Save()
	}
	return nil
}

func AddProvider(provider Provider) error {
	if cfg == nil {
		return nil
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.Providers = append(cfg.Providers, provider)
	return cfg.Save()
}

func UpdateProvider(index int, provider Provider) error {
	if cfg == nil {
		return nil
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if index >= 0 && index < len(cfg.Providers) {
		cfg.Providers[index] = provider
		return cfg.Save()
	}
	return nil
}

func DeleteProvider(index int) error {
	if cfg == nil {
		return nil
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if index >= 0 && index < len(cfg.Providers) {
		cfg.Providers = append(cfg.Providers[:index], cfg.Providers[index+1:]...)
		if cfg.ActiveProvider >= len(cfg.Providers) {
			cfg.ActiveProvider = 0
		}
		return cfg.Save()
	}
	return nil
}

func GetActiveProviderIndex() int {
	if cfg == nil {
		return 0
	}
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	return cfg.ActiveProvider
}

func GetModels() []string {
	provider := GetActiveProvider()
	if provider == nil {
		return []string{}
	}
	return provider.Models
}

func GetTrayMode() bool {
	if cfg == nil {
		return false
	}
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	return cfg.TrayMode
}

func SetTrayMode(enabled bool) error {
	if cfg == nil {
		return nil
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.TrayMode = enabled
	return cfg.Save()
}
