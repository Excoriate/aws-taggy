package output

import (
	"fmt"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v3"
)

// WriteToClipboard copies the validation result to the clipboard in YAML format
func WriteToClipboard(data interface{}) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %w", err)
	}

	err = clipboard.WriteAll(string(content))
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
