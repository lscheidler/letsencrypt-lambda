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
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/acme"

	"github.com/lscheidler/letsencrypt-lambda/account/certificate"
	"github.com/lscheidler/letsencrypt-lambda/account/certificate/privatekey"
	"github.com/lscheidler/letsencrypt-lambda/account/registration"
	"github.com/lscheidler/letsencrypt-lambda/provider"
)

const (
	// staging
	//DirectoryURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	DirectoryURL = "https://acme-v02.api.letsencrypt.org/directory"
)

type Accounts []Account

type Account struct {
	Certificates     map[string]*certificate.Certificate `json:"certificates"`
	Changed          bool                                `json:"-"`
	ClientPassphrase *string                             `json:"-"`
	Domains          []string                            `json:"-"`
	Email            *string                             `json:"-"`
	Registration     *registration.RegistrationCrypt     `json:"registration"`
	client           *acme.Client
	provider         *provider.Provider
}

func New(email *string, domains []string, provider *provider.Provider) *Account {
	return &Account{
		Certificates: map[string]*certificate.Certificate{},
		Changed:      false,
		Domains:      domains,
		Email:        email,
		Registration: &registration.RegistrationCrypt{},
		provider:     provider,
	}
}

func (a *Account) Create() error {
	a.Changed = true
	a.Registration.Key = privatekey.New()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	a.client = &acme.Client{
		Key:          a.Registration.Key.Signer(),
		DirectoryURL: DirectoryURL,
	}

	acmeAccount := &acme.Account{Contact: []string{"mailto:" + *a.Email}}
	reg, err := a.client.Register(ctx, acmeAccount, acme.AcceptTOS)
	if err != nil {
		return err
	}

	a.Registration.Contact = reg.Contact
	a.Registration.URI = reg.URI
	a.Registration.OrdersURL = reg.OrdersURL
	a.Registration.Status = reg.Status
	return nil
}

func (a *Account) CreateOrRenewCertificates() error {
	if a.client == nil {
		return fmt.Errorf("acme.Client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var cert *certificate.Certificate
	if cert = a.Certificates[fmt.Sprintf("%v", a.Domains)]; cert == nil {
		cert = certificate.New(a.Domains)
		a.Certificates[fmt.Sprintf("%v", a.Domains)] = cert
	} else {
		now := time.Now()
		if duration := cert.NotAfter.Sub(now).Hours(); duration >= 30*24 {
			// if NotAfter is >= 30 days away, skip renew
			log.Printf("The certificate is valid for %d days. Skipping renewal.", int(duration/24))
			return nil
		}
	}

	csr, err := cert.Request()
	if err != nil {
		return err
	}

	dir, err := a.client.Discover(ctx)
	if err != nil {
		return err
	}

	if dir.OrderURL == "" {
		log.Fatal("Pre-RFC legacy CA not supported")
	}

	// verify domain
	order, err := a.verify(ctx)
	if err != nil {
		return err
	}

	der, _, err := a.client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return err
	}

	err = cert.Add(der)
	if err != nil {
		return err
	}
	a.Changed = true
	return nil
}

func (a *Account) verify(ctx context.Context) (*acme.Order, error) {
	var order *acme.Order
	var err error

	// get AuthorizeOrder for domain
	log.Println("AuthorizeOrder", a.Domains)
	if order, err = a.client.AuthorizeOrder(ctx, acme.DomainIDs(a.Domains...)); err != nil {
		return nil, err
	}

	// Remove all hanging authorizations to reduce rate limit quotas
	// after we're done.
	defer func(urls []string) {
		go a.deactivatePendingAuthz(urls)
	}(order.AuthzURLs)

	// https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L778
	switch order.Status {
	case acme.StatusReady:
		// Already authorized.
		return order, nil
	case acme.StatusPending:
		// Continue normal Order-based flow.
	default:
		return nil, fmt.Errorf("invalid new order status %q; order URL: %q", order.Status, order.URI)
	}

	// Satisfy all pending authorizations.
	// https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L789
	for _, zurl := range order.AuthzURLs {
		log.Println("GetAuthorization", zurl)
		z, err := a.client.GetAuthorization(ctx, zurl)
		if err != nil {
			return nil, err
		}
		if z.Status != acme.StatusPending {
			// We are interested only in pending authorizations.
			continue
		}

		log.Println("Create challenge record for", z.Identifier.Value)
		challenge := pickChallenge("dns-01", z.Challenges)
		var token string
		if token, err = a.client.DNS01ChallengeRecord(challenge.Token); err != nil {
			return nil, err
		}

		// challenge fulfilment
		(*a.provider).CreateChallenge("_acme-challenge."+z.Identifier.Value+".", token)

		log.Println("Accept")
		if _, err = a.client.Accept(ctx, challenge); err != nil {
			return nil, err
		}
		log.Println("WaitAuthorization")
		if _, err = a.client.WaitAuthorization(ctx, z.URI); err != nil {
			return nil, err
		}

		(*a.provider).RemoveChallenge("_acme-challenge."+z.Identifier.Value+".", token)
	}

	log.Println("WaitOrder")
	if order, err = a.client.WaitOrder(ctx, order.URI); err != nil {
		return nil, err
	}

	return order, nil
}

// deactivatePendingAuthz relinquishes all authorizations identified by the elements
// of the provided uri slice which are in "pending" state.
// It ignores revocation errors.
//
// deactivatePendingAuthz takes no context argument and instead runs with its own
// "detached" context because deactivations are done in a goroutine separate from
// that of the main issuance or renewal flow.
// see: https://github.com/golang/crypto/blob/5c72a883971a4325f8c62bf07b6d38c20ea47a6a/acme/autocert/autocert.go#L857
func (a *Account) deactivatePendingAuthz(uri []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for _, u := range uri {
		z, err := a.client.GetAuthorization(ctx, u)
		if err == nil && z.Status == acme.StatusPending {
			a.client.RevokeAuthorization(ctx, u)
		}
	}
}
