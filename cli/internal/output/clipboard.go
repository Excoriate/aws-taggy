package output

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v3"
)

// WriteToClipboard copies the validation result to the clipboard in YAML format
func WriteToClipboard(data interface{}) error {
	// Use a more robust YAML marshaling approach
	content, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %w", err)
	}

	// Trim any trailing newlines to ensure clean clipboard content
	yamlString := strings.TrimSpace(string(content))

	err = clipboard.WriteAll(yamlString)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
