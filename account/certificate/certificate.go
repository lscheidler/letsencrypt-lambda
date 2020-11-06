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

package certificate

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"time"

	"github.com/lscheidler/letsencrypt-lambda/account/certificate/privatekey"
	"github.com/lscheidler/letsencrypt-lambda/crypto"
)

type Certificate struct {
	Domains       []string `json:"domains"`
	CertUrl       *string  `json:"certUrl"`
	CertStableUrl *string  `json:"certStableUrl"`

	CreatedAt time.Time `json:"createdAt"`
	NotAfter  time.Time `json:"notAfter"`
	Pem       []byte    `json:"pem"`

	KeyCreatedAt time.Time              `json:"privateKeyCreatedAt"`
	Key          *privatekey.PrivateKey `json:"privateKey"`
}

func New(domains []string) *Certificate {
	return &Certificate{
		Domains:      domains,
		Key:          privatekey.New(),
		KeyCreatedAt: time.Now(),
	}
}

func (c *Certificate) Add(data [][]byte) error {
	c.CreatedAt = time.Now()
	leaf, err := c.ValidCert(data, c.CreatedAt)
	if err != nil {
		return err
	}
	tlsCertificate := c.tlscert(data, leaf)
	c.NotAfter = leaf.NotAfter

	c.Pem, err = generatePem(tlsCertificate)
	if err != nil {
		return err
	}
	return nil
}

// certRequest generates a CSR for the given common name cn and optional SANs.
// see: https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L1137
func (c *Certificate) Request() ([]byte, error) {
	req := &x509.CertificateRequest{
		Subject:         pkix.Name{CommonName: c.Domains[0]}, // cn
		DNSNames:        c.Domains[1:len(c.Domains)],         // san
		ExtraExtensions: []pkix.Extension{},
	}
	return x509.CreateCertificateRequest(rand.Reader, req, c.Key.Signer())
}

// validCert parses a cert chain provided as der argument and verifies the leaf and der[0]
// correspond to the private key, the domain and key type match, and expiration dates
// are valid. It doesn't do any revocation checking.
//
// The returned value is the verified leaf cert.
// see: https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L1177
func (c *Certificate) ValidCert(der [][]byte, now time.Time) (leaf *x509.Certificate, err error) {
	// parse public part(s)
	var n int
	for _, b := range der {
		n += len(b)
	}
	pub := make([]byte, n)
	n = 0
	for _, b := range der {
		n += copy(pub[n:], b)
	}
	x509Cert, err := x509.ParseCertificates(pub)
	if err != nil || len(x509Cert) == 0 {
		return nil, errors.New("no public key found")
	}
	// verify the leaf is not expired and matches the domain name
	leaf = x509Cert[0]
	if now.Before(leaf.NotBefore) {
		return nil, errors.New("certificate is not valid yet")
	}
	if now.After(leaf.NotAfter) {
		return nil, errors.New("expired certificate")
	}
	if err := leaf.VerifyHostname(c.Domains[0]); err != nil {
		return nil, err
	}
	// ensure the leaf corresponds to the private key and matches the certKey type
	switch pub := leaf.PublicKey.(type) {
	case *ecdsa.PublicKey:
		prv := ecdsa.PrivateKey(*c.Key)
		if pub.X.Cmp(prv.X) != 0 || pub.Y.Cmp(prv.Y) != 0 {
			return nil, errors.New("private key does not match public key")
		}
	default:
		return nil, errors.New("unknown public key algorithm")
	}
	return leaf, nil
}

func (c *Certificate) tlscert(cert [][]byte, leaf *x509.Certificate) *tls.Certificate {
	return &tls.Certificate{
		PrivateKey:  c.Key.Signer(),
		Certificate: cert,
		Leaf:        leaf,
	}
}

func generatePem(tlscert *tls.Certificate) ([]byte, error) {
	var buf bytes.Buffer
	crypto.EncodeECDSAKey(&buf, tlscert.PrivateKey.(*ecdsa.PrivateKey))

	// public
	// see: https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L536
	for _, b := range tlscert.Certificate {
		pb := &pem.Block{Type: "CERTIFICATE", Bytes: b}
		if err := pem.Encode(&buf, pb); err != nil {
			return buf.Bytes(), err
		}
	}
	return buf.Bytes(), nil
}
