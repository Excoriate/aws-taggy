package normaliser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeServiceName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"S3", "s3"},
		{"s3", "s3"},
		{"EC2", "ec2"},
		{"ec2", "ec2"},
		{" S3 ", "s3"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := NormalizeServiceName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
