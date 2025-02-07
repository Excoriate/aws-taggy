package output

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"gopkg.in/yaml.v3"
)

type ConfigurationWriter struct {
	Config *configuration.TaggyScanConfig
	File   string
}

type DocumentationWriter struct {
	File string
}

func NewConfigurationWriter() *ConfigurationWriter {
	return &ConfigurationWriter{
		Config: &configuration.TaggyScanConfig{},
	}
}

func NewDocumentationWriter() *DocumentationWriter {
	return &DocumentationWriter{
		File: "how-to-customize-aws-taggy-configuration.md",
	}
}

func (w *ConfigurationWriter) WriteConfiguration(file string, overwrite bool) error {
	// Set default configuration
	w.SetDefaultConfig()

	// Validate the file path
	if file == "" {
		return fmt.Errorf("output file path cannot be empty")
	}

	// Check if file exists and overwrite is not allowed
	if !overwrite {
		if _, err := os.Stat(file); err == nil {
			return fmt.Errorf("configuration file already exists at %s. Use the -f flag to overwrite", file)
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory for configuration file: %w", err)
	}

	// Marshal the configuration to YAML
	yamlData, err := yaml.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to YAML: %w", err)
	}

	// Write the YAML to file
	if err := os.WriteFile(file, yamlData, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

func (w *ConfigurationWriter) SetDefaultConfig() {
	w.Config = configuration.DefaultConfiguration()
}

func (w *DocumentationWriter) WriteDocumentation(configFile string) error {
	// Validate the file path
	if configFile == "" {
		return fmt.Errorf("configuration file path cannot be empty")
	}

	// Generate documentation file path
	docFile := configuration.GenerateDocumentationFilename(configFile)

	// Ensure directory exists
	dir := filepath.Dir(docFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory for documentation file: %w", err)
	}

	// Get default documentation content
	content := configuration.DefaultDocumentation()

	// Write the documentation to file
	if err := os.WriteFile(docFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write documentation file: %w", err)
	}

	return nil
}
