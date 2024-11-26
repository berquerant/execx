//go:build tools
// +build tools

package main

import (
	_ "github.com/berquerant/goconfig"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
