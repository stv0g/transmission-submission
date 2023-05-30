// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type VAPIDKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func (kp *VAPIDKeyPair) Load(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	defer f.Close()

	return json.NewDecoder(f).Decode(kp)
}

func (kp *VAPIDKeyPair) Save(fn string) error {
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	return json.NewEncoder(f).Encode(kp)
}

func (kp *VAPIDKeyPair) Create() error {
	sk, pk, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return err
	}

	kp.PrivateKey = sk
	kp.PublicKey = pk

	return nil
}

func LoadOrCreateVAPIDKeyPair(fn string) (VAPIDKeyPair, error) {
	var vapid VAPIDKeyPair

	if err := vapid.Load(fn); err != nil {
		if err := vapid.Create(); err != nil {
			return vapid, err
		}

		if err := vapid.Save(fn); err != nil {
			return vapid, err
		}
	}

	return vapid, nil
}
