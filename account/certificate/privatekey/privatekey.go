/*
Copyright 2020 Lars Eric Scheidler

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package privatekey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"log"
)

type PrivateKey ecdsa.PrivateKey

func New() *PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	privateKey := PrivateKey(*privKey)
	return &privateKey
}

func (p *PrivateKey) UnmarshalJSON(b []byte) error {
	var j []byte
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	} else {
		var ep *ecdsa.PrivateKey
		if ep, err = x509.ParseECPrivateKey(j); err != nil {
			return err
		} else {
			*p = PrivateKey(*ep)
		}
	}

	return nil
}

func (p *PrivateKey) MarshalJSON() ([]byte, error) {
	ep := ecdsa.PrivateKey(*p)
	b, err := x509.MarshalECPrivateKey(&ep)
	if err != nil {
		return nil, err
	}

	return json.Marshal(b)
}

func (p *PrivateKey) Signer() crypto.Signer {
	ep := ecdsa.PrivateKey(*p)
	return crypto.Signer(&ep)
}
