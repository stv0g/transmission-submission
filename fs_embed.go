//go:build embed

package main

import "embed"

var (
	//go:embed assets
	assetFiles embed.FS

	//go:embed templates
	tplFiles embed.FS
)
