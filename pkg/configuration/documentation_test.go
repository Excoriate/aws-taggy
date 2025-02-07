package configuration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDocumentation(t *testing.T) {
	doc := DefaultDocumentation()

	// Test that the documentation contains essential sections
	essentialSections := []string{
		"# AWS Taggy Configuration Guide",
		"## Configuration Structure",
		"### Version",
		"### AWS Configuration",
		"### Global Settings",
		"### Resource-Specific Configurations",
		"### Compliance Levels",
		"### Tag Validation Rules",
		"### Notifications",
		"## Best Practices",
	}

	for _, section := range essentialSections {
		assert.True(t, strings.Contains(doc, section), "Documentation should contain section: %s", section)
	}

	// Test that the documentation contains important configuration elements
	configElements := []string{
		"minimum_required_tags",
		"max_tags",
		"required_tags",
		"forbidden_tags",
		"specific_tags",
		"compliance_level",
		"batch_size",
	}

	for _, element := range configElements {
		assert.True(t, strings.Contains(doc, element), "Documentation should contain configuration element: %s", element)
	}
}

func TestGenerateDocumentationFilename(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic config file",
			input:    "config.yaml",
			expected: "config.yaml.md",
		},
		{
			name:     "with path",
			input:    "/path/to/config.yaml",
			expected: "/path/to/config.yaml.md",
		},
		{
			name:     "empty string",
			input:    "",
			expected: ".md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateDocumentationFilename(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
