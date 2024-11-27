package inspector

import (
	"time"

	"github.com/Excoriate/aws-taggy/pkg/constants"
)

type Resource interface {
	GetTags() map[string]string
	GetRegion() string
	GetType() string
}

type ResourceMetadata struct {
	// Basic resource identification
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Provider     string            `json:"provider"`
	Region       string            `json:"region"`
	AccountID    string            `json:"account_id"`
	Tags         map[string]string `json:"tags"`
	DiscoveredAt time.Time         `json:"discovered_at"`

	// Extended information
	Details struct {
		ARN        string                 `json:"arn,omitempty"`
		Name       string                 `json:"name,omitempty"`
		Status     string                 `json:"status,omitempty"`
		Properties map[string]interface{} `json:"properties,omitempty"`
		Compliance struct {
			IsCompliant bool      `json:"is_compliant"`
			Violations  []string  `json:"violations,omitempty"`
			LastCheck   time.Time `json:"last_check"`
		} `json:"compliance"`
	} `json:"details"`

	// Store complete API response
	RawResponse interface{} `json:"raw_response,omitempty"`
}

type BaseResource struct {
	Type   string
	Region string
	Tags   map[string]string
}

func (r *BaseResource) GetType() string {
	return r.Type
}

func (r *BaseResource) GetRegion() string {
	if r.Region == "" {
		return constants.DefaultAWSRegion
	}
	return r.Region
}

func (r *BaseResource) GetTags() map[string]string {
	if r.Tags == nil {
		r.Tags = make(map[string]string)
	}
	return r.Tags
}

func NewResourceType(resourceType string) Resource {
	return &BaseResource{
		Type: resourceType,
		Tags: make(map[string]string),
	}
}
