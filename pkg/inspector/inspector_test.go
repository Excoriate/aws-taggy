package inspector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResourceMetadata(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		resource ResourceMetadata
		expected bool
	}{
		{
			name: "Compliant Resource",
			resource: ResourceMetadata{
				ID:        "test-resource-1",
				Type:      "s3",
				Provider:  "aws",
				Region:    "us-east-1",
				AccountID: "123456789",
				Tags: map[string]string{
					"environment": "production",
					"owner":       "team-a",
				},
				DiscoveredAt: time.Now(),
				Details: struct {
					ARN        string                 `json:"arn,omitempty"`
					Name       string                 `json:"name,omitempty"`
					Status     string                 `json:"status,omitempty"`
					Properties map[string]interface{} `json:"properties,omitempty"`
					Compliance struct {
						IsCompliant bool      `json:"is_compliant"`
						Violations  []string  `json:"violations,omitempty"`
						LastCheck   time.Time `json:"last_check"`
					} `json:"compliance"`
				}{
					Name:   "test-bucket",
					Status: "active",
					Compliance: struct {
						IsCompliant bool      `json:"is_compliant"`
						Violations  []string  `json:"violations,omitempty"`
						LastCheck   time.Time `json:"last_check"`
					}{
						IsCompliant: true,
						LastCheck:   time.Now(),
					},
				},
			},
			expected: true,
		},
		{
			name: "Non-Compliant Resource",
			resource: ResourceMetadata{
				ID:        "test-resource-2",
				Type:      "ec2",
				Provider:  "aws",
				Region:    "us-west-2",
				AccountID: "987654321",
				Tags: map[string]string{
					"environment": "staging",
				},
				DiscoveredAt: time.Now(),
				Details: struct {
					ARN        string                 `json:"arn,omitempty"`
					Name       string                 `json:"name,omitempty"`
					Status     string                 `json:"status,omitempty"`
					Properties map[string]interface{} `json:"properties,omitempty"`
					Compliance struct {
						IsCompliant bool      `json:"is_compliant"`
						Violations  []string  `json:"violations,omitempty"`
						LastCheck   time.Time `json:"last_check"`
					} `json:"compliance"`
				}{
					Name:   "test-instance",
					Status: "running",
					Compliance: struct {
						IsCompliant bool      `json:"is_compliant"`
						Violations  []string  `json:"violations,omitempty"`
						LastCheck   time.Time `json:"last_check"`
					}{
						IsCompliant: false,
						Violations:  []string{"missing_tags"},
						LastCheck:   time.Now(),
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.resource.Details.Compliance.IsCompliant)
		})
	}
}

func TestBaseResource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		resourceType   string
		region         string
		expectedRegion string
	}{
		{
			name:           "Resource with Specified Region",
			resourceType:   "s3",
			region:         "us-west-2",
			expectedRegion: "us-west-2",
		},
		{
			name:           "Resource with Empty Region",
			resourceType:   "ec2",
			region:         "",
			expectedRegion: "us-east-1", // Default region
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource := &BaseResource{
				Type:   tc.resourceType,
				Region: tc.region,
				Tags:   make(map[string]string),
			}

			assert.Equal(t, tc.resourceType, resource.GetType())
			assert.Equal(t, tc.expectedRegion, resource.GetRegion())
			assert.NotNil(t, resource.GetTags())
		})
	}
}
