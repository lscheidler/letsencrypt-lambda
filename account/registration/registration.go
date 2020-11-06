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

package registration

import (
	"encoding/json"
	"log"

	"github.com/lscheidler/letsencrypt-lambda/account/certificate/privatekey"
	"github.com/lscheidler/letsencrypt-lambda/crypto"
	"github.com/lscheidler/letsencrypt-lambda/helper"
	"github.com/lscheidler/letsencrypt-lambda/secrets"
)

type Registration struct {
	Key       *privatekey.PrivateKey `json:"privateKey"`
	Status    string                 `json:"status"`
	Contact   []string               `json:"contact"`
	URI       string                 `json:"uri"`
	OrdersURL string                 `json:"ordersURL"`
}

type RegistrationCrypt Registration

func (rc *RegistrationCrypt) UnmarshalJSON(b []byte) error {
	var jsonData []byte
	var plaintext []byte
	var err error
	var issuerPassphrase *string

	if issuerPassphrase = getIssuerPassphrase(); issuerPassphrase == nil {
		return nil
	}

	if err = json.Unmarshal(b, &jsonData); err != nil {
		return err
	}

	if plaintext, err = crypto.Decrypt(jsonData, []byte(*issuerPassphrase)); err != nil {
		return err
	}

	r := Registration(*rc)
	if err := json.Unmarshal(plaintext, &r); err != nil {
		return err
	}
	*rc = RegistrationCrypt(r)
	return nil
}

func (rc *RegistrationCrypt) MarshalJSON() ([]byte, error) {
	var ciphertext []byte
	var issuerPassphrase *string

	if issuerPassphrase = getIssuerPassphrase(); issuerPassphrase == nil {
		return ciphertext, nil
	}

	r := Registration(*rc)
	plaintext, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	if ciphertext, err = crypto.Encrypt(plaintext, []byte(*issuerPassphrase)); err != nil {
		return nil, err
	}

	return json.Marshal(ciphertext)
}

func getIssuerPassphrase() *string {
	if issuerPassphrase := helper.Getenv("ISSUER_PASSPHRASE"); issuerPassphrase == nil {
		if issuerPassphraseSecretsArn := helper.Getenv("ISSUER_PASSPHRASE_SECRET_ARN"); issuerPassphraseSecretsArn == nil {
			log.Println("Environment variable ISSUER_PASSPHRASE and ISSUER_PASSPHRASE_SECRET_ARN not found. For issuer en/decryption one is required. Not required in client mode.")
			return nil
		} else {
			if s := secrets.GetSecret(issuerPassphraseSecretsArn); s != nil {
				return s
			} else {
				log.Printf("Secret %s not found", *issuerPassphraseSecretsArn)
				return nil
			}
		}
	} else {
		return issuerPassphrase
	}
}
