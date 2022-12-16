//go:build !embed

package main

import (
	"os"
)

var (
	assetFiles = os.DirFS(".")
	tplFiles   = assetFiles
)
