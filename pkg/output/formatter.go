package output

import (
	"encoding/json"
	"fmt"
)

// ToJSON converts the input to a JSON byte slice
func (f *Formatter) ToJSON(data interface{}) ([]byte, error) {
	// If the output format is not JSON, convert to JSON
	jsonOutput, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return jsonOutput, nil
}
