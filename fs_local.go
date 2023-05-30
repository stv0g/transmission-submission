// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !embed

package main

import (
	"os"
)

var (
	assetFiles = os.DirFS(".")
	tplFiles   = assetFiles
)
