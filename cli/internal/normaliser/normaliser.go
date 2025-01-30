package normaliser

import (
	"strings"
)

// NormalizeServiceName converts service names to a consistent lowercase format
// Handles variations like "S3", "s3", "EC2", "ec2"
func NormalizeServiceName(serviceName string) string {
	return strings.ToLower(strings.TrimSpace(serviceName))
}
