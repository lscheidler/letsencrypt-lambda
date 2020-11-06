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

package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	defaultRegion = "eu-central-1"
)

func GetAwsSession() (*session.Session, *aws.Config) {
	conf := &aws.Config{Region: aws.String(defaultRegion)}
	if region := os.Getenv("REGION"); len(region) > 0 {
		log.Println("getAwsSession: set region to ", region)
		conf.Region = aws.String(region)
	}

	sess := session.Must(session.NewSession())
	if role := os.Getenv("ASSUME_ROLE"); len(role) > 0 {
		log.Println("getAwsSession: assume role ", role)
		creds := stscreds.NewCredentials(sess, role)
		conf.Credentials = creds
	}
	return sess, conf
}
