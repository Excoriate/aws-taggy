package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppConstants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "App Name",
			constant: AppName,
			expected: "aws-taggy",
		},
		{
			name:     "App Description",
			constant: AppDescription,
			expected: "A powerful CLI to inspect and manage AWS resources tags",
		},
		{
			name:     "Supported Config Version",
			constant: SupportedConfigVersion,
			expected: "1.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.constant)
		})
	}
}

func TestAWSConstants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "Default AWS Region",
			constant: DefaultAWSRegion,
			expected: "us-east-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.constant)
		})
	}
}
