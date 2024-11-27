package output

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
)

// CopyToClipboard converts data to JSON and copies it to the clipboard
func CopyToClipboard(data interface{}) error {
	// Convert to JSON
	jsonOutput, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(string(jsonOutput)); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
