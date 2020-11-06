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

package main

import (
	"flag"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/lscheidler/letsencrypt-lambda/account"
	"github.com/lscheidler/letsencrypt-lambda/dynamodb"
	"github.com/lscheidler/letsencrypt-lambda/helper"
	"github.com/lscheidler/letsencrypt-lambda/provider"
	"github.com/lscheidler/letsencrypt-lambda/provider/dns/route53"
)

type env struct {
	awsHostedZoneId   *string
	debug             bool
	domains           []string
	dynamodbTableName *string
	email             *string
}

func loadEnv() *env {
	env := &env{}

	env.debug = helper.GetenvBool("DEBUG")

	if email := helper.Getenv("EMAIL"); email != nil {
		env.email = email
	} else {
		log.Println("Environment variable EMAIL not found.")
		return nil
	}

	if domainsStr := helper.Getenv("DOMAINS"); domainsStr != nil {
		domains := strings.Split(*domainsStr, ",")
		for index := range domains {
			env.domains = append(env.domains, domains[index])
		}
	} else {
		log.Println("Environment variable DOMAINS not found.")
		return nil
	}

	if awsHostedZoneId := helper.Getenv("AWS_HOSTED_ZONE_ID"); awsHostedZoneId != nil {
		env.awsHostedZoneId = awsHostedZoneId
	} else {
		log.Println("Environment variable AWS_HOSTED_ZONE_ID not found.")
		return nil
	}

	if dynamodbTableName := helper.Getenv("DYNAMODB_TABLE_NAME"); dynamodbTableName != nil {
		env.dynamodbTableName = dynamodbTableName
	}

	if issuerPassphrase := helper.Getenv("ISSUER_PASSPHRASE"); issuerPassphrase == nil {
		if issuerPassphraseSecretsArn := helper.Getenv("ISSUER_PASSPHRASE_SECRET_ARN"); issuerPassphraseSecretsArn == nil {
			log.Fatal("Environment variable ISSUER_PASSPHRASE and ISSUER_PASSPHRASE_SECRET_ARN not found. One of these environment variables must be set.")
		}
	}

	if clientPassphrase := helper.Getenv("CLIENT_PASSPHRASE"); clientPassphrase == nil {
		if clientPassphraseSecretsArn := helper.Getenv("CLIENT_PASSPHRASE_SECRET_ARN"); clientPassphraseSecretsArn == nil {
			log.Fatal("Environment variable CLIENT_PASSPHRaASE and CLIENT_PASSPHRASE_SECRET_ARN not found. One of these environment variables must be set.")
		}
	}

	return env
}

func main() {
	local := flag.Bool("local", false, "run lambda function localy")
	flag.Parse()

	if *local {
		if err := HandleRequest(); err != nil {
			log.Fatal(err)
		}
	} else {
		lambda.Start(HandleRequest)
	}
}

func HandleRequest() error {
	env := loadEnv()

	// Load provider
	route53 := route53.New(env.awsHostedZoneId)
	p := provider.Provider(route53)

	account := account.New(env.email, env.domains, &p)
	dynamodb := dynamodb.New(env.dynamodbTableName)

	if err := dynamodb.CreateOrLoadAccount(account); err != nil {
		return err
	}

	if err := account.CreateOrRenewCertificates(); err != nil {
		return err
	}

	if err := dynamodb.Update(account); err != nil {
		return err
	}
	return nil
}
