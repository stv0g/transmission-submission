// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestDecodeLabels(t *testing.T) {
	m := decodeLabels([]string{
		"x-hello=world",
		"x-hello=world2",
		"x-bla=blub",
	})

	if len(m) != 2 {
		t.Fail()
	}

	if v, ok := m["hello"]; !ok || v != "world2" {
		t.Fail()
	}

	if v, ok := m["bla"]; !ok || v != "blub" {
		t.Fail()
	}
}
