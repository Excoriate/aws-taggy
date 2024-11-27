package output

// ValidationResult represents the structured output of configuration validation
type ValidationResult struct {
	Status    string   `json:"status" yaml:"status"`
	File      string   `json:"file" yaml:"file"`
	Valid     bool     `json:"valid" yaml:"valid"`
	Errors    []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings  []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Version   string   `json:"version" yaml:"version"`
	Resources struct {
		Total    int      `json:"total" yaml:"total"`
		Enabled  int      `json:"enabled" yaml:"enabled"`
		Services []string `json:"services" yaml:"services"`
	} `json:"resources" yaml:"resources"`
	GlobalConfig struct {
		Enabled            bool     `json:"enabled" yaml:"enabled"`
		MinRequiredTags    int      `json:"min_required_tags" yaml:"min_required_tags"`
		RequiredTags       []string `json:"required_tags" yaml:"required_tags"`
		ForbiddenTags      []string `json:"forbidden_tags" yaml:"forbidden_tags"`
		ComplianceLevel    string   `json:"compliance_level" yaml:"compliance_level"`
		BatchSize          int      `json:"batch_size" yaml:"batch_size"`
		NotificationsSetup bool     `json:"notifications_setup" yaml:"notifications_setup"`
	} `json:"global_config" yaml:"global_config"`
	ComplianceLevels  []string            `json:"compliance_levels" yaml:"compliance_levels"`
	ComplianceResults []*ComplianceResult `json:"compliance_results,omitempty" yaml:"compliance_results,omitempty"`
	ComplianceSummary *ComplianceSummary  `json:"compliance_summary,omitempty" yaml:"compliance_summary,omitempty"`
}
