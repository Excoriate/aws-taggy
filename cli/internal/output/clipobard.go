package output

import (
	"fmt"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v3"
)

// CopyToClipboard converts data to YAML and copies it to the clipboard
func CopyToClipboard(data interface{}) error {
	// Convert to YAML
	yamlOutput, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(string(yamlOutput)); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
