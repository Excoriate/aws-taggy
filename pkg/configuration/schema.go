package configuration

import (
	_ "embed"
)

//go:embed schema/tag-compliance-schema.json
var tagComplianceSchema string
