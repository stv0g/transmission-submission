// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build embed

package main

import "embed"

var (
	//go:embed assets
	assetFiles embed.FS

	//go:embed templates
	tplFiles embed.FS
)
