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

package account

import (
	"encoding/json"
	"log"

	"golang.org/x/crypto/acme"

	"github.com/lscheidler/letsencrypt-lambda/crypto"
	"github.com/lscheidler/letsencrypt-lambda/helper"
	"github.com/lscheidler/letsencrypt-lambda/secrets"
)

type AccountCrypt Account

func (ac *AccountCrypt) UnmarshalJSON(b []byte) error {
	var jsonData []byte
	var plaintext []byte
	var err error

	clientPassphrase := getClientPassphrase()

	if err = json.Unmarshal(b, &jsonData); err != nil {
		return err
	}

	if plaintext, err = crypto.Decrypt(jsonData, []byte(*clientPassphrase)); err != nil {
		return err
	}

	a := Account(*ac)
	if err := json.Unmarshal(plaintext, &a); err != nil {
		return err
	}
	*ac = AccountCrypt(a)

	if ac.Registration.Key != nil {
		ac.client = &acme.Client{
			Key:          ac.Registration.Key.Signer(),
			DirectoryURL: DirectoryURL,
		}
	}
	return nil
}

func (ac *AccountCrypt) MarshalJSON() ([]byte, error) {
	var ciphertext []byte

	clientPassphrase := getClientPassphrase()

	a := Account(*ac)
	plaintext, err := json.Marshal(&a)
	if err != nil {
		return nil, err
	}

	if ciphertext, err = crypto.Encrypt(plaintext, []byte(*clientPassphrase)); err != nil {
		return nil, err
	}

	return json.Marshal(ciphertext)
}

// https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L831
func pickChallenge(typ string, chal []*acme.Challenge) *acme.Challenge {
	for _, c := range chal {
		if c.Type == typ {
			return c
		}
	}
	return nil
}

func getClientPassphrase() *string {
	if clientPassphrase := helper.Getenv("CLIENT_PASSPHRASE"); clientPassphrase == nil {
		if clientPassphraseSecretsArn := helper.Getenv("CLIENT_PASSPHRASE_SECRET_ARN"); clientPassphraseSecretsArn == nil {
			log.Fatal("Environment variable CLIENT_PASSPHRASE and CLIENT_PASSPHRASE_SECRET_ARN not found. One of these environment variables must be set.")
			return nil
		} else {
			if s := secrets.GetSecret(clientPassphraseSecretsArn); s != nil {
				return s
			} else {
				log.Fatalf("Secret %s not found", *clientPassphraseSecretsArn)
				return nil
			}
		}
	} else {
		return clientPassphrase
	}
}
