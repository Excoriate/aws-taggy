package normaliser

import (
	"strings"
)

// NormalizeServiceName converts service names to a consistent lowercase format
// Handles variations like "S3", "s3", "EC2", "ec2"
func NormalizeServiceName(serviceName string) string {
	return strings.ToLower(strings.TrimSpace(serviceName))
}

// NormalizeOutputFormat converts output format to a consistent lowercase format
// Handles variations like "JSON", "json", "YAML", "yml"
func NormalizeOutputFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "table":
		return "table"
	default:
		return "table" // Default to table if unrecognized
	}
}
