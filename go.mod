module github.com/berquerant/execx

go 1.24.4

require (
	github.com/stretchr/testify v1.10.0
	golang.org/x/sync v0.16.0
)

require (
	github.com/berquerant/goconfig v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/telemetry v0.0.0-20240522233618-39ace7a40ae7 // indirect
	golang.org/x/tools v0.29.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool (
	github.com/berquerant/goconfig
	golang.org/x/vuln/cmd/govulncheck
)
