module github.com/Excoriate/aws-taggy/cli

go 1.23.3

require (
	github.com/Excoriate/aws-taggy v0.0.0
	github.com/alecthomas/kong v1.4.0
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace github.com/Excoriate/aws-taggy => ../
