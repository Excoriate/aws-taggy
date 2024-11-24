package scannconfig

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ConfigLoader handles loading configuration files
type ConfigLoader struct {
	config *TaggyScanConfig
}

// NewTaggyScanConfigLoader creates a new ConfigLoader instance
func NewTaggyScanConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

// LoadConfig loads a configuration file from the specified path
// LoadConfig performs the following steps:
// 1. Validate the configuration file path and existence
// 2. Parse the YAML configuration
// 3. Validate the parsed configuration structure
//
// Parameters:
//   - configPath: Full path to the configuration file
//
// Returns:
//   - *TaggyScanConfig: Fully loaded and validated configuration
//   - error: Any error encountered during loading or validation
func (l *ConfigLoader) LoadConfig(configPath string) (*TaggyScanConfig, error) {
	// Validate file path and existence
	fileValidator, err := NewConfigFileValidator(configPath)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration file path: %w", err)
	}

	// Perform file validation
	if err := fileValidator.Validate(); err != nil {
		return nil, fmt.Errorf("configuration file validation failed: %w", err)
	}

	// Read file contents
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse YAML
	parsedCfg := &TaggyScanConfig{}
	if err := yaml.Unmarshal(fileContent, parsedCfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Normalize AWS configuration
	NormalizeAWSConfig(&parsedCfg.AWS)

	// Validate configuration content
	configValidator, err := NewConfigValidator(parsedCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration validator: %w", err)
	}

	// Perform comprehensive configuration validation
	if err := configValidator.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Store the loaded configuration
	l.config = parsedCfg

	return parsedCfg, nil
}

// GetLoadedConfig returns the currently loaded configuration
// Returns nil if no configuration has been loaded
func (l *ConfigLoader) GetLoadedConfig() *TaggyScanConfig {
	return l.config
}

// CompilePatternRules compiles regex patterns for tag validation
// This method can be moved to the ConfigValidator if it makes more sense
func (l *ConfigLoader) CompilePatternRules() error {
	if l.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	// Initialize compiledRules if not already initialized
	if l.config.TagValidation.compiledRules == nil {
		l.config.TagValidation.compiledRules = make(map[string]*regexp.Regexp)
	}

	// Compile pattern rules
	for tagName, pattern := range l.config.TagValidation.PatternRules {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern for tag %s: %w", tagName, err)
		}
		l.config.TagValidation.compiledRules[tagName] = compiled
	}

	return nil
}

// GetComplianceLevelRequirements returns the compliance level requirements for the specified level
// Returns nil if the level is not found
func (l *ConfigLoader) GetComplianceLevelRequirements(level string) (*ComplianceLevel, error) {
	complianceLevel, exists := l.config.ComplianceLevels[level]
	if !exists {
		return nil, fmt.Errorf("compliance level %s not found", level)
	}
	return &complianceLevel, nil
}
