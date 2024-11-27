package cmd

// ComplianceCmd represents the compliance command group
type ComplianceCmd struct {
	Check CheckCmd `cmd:"" help:"Check AWS resource tag compliance"`
}

// Run is a no-op method to satisfy the Kong command interface
func (c *ComplianceCmd) Run() error {
	return nil
}
