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

package route53

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"

	awshelper "github.com/lscheidler/letsencrypt-lambda/helper/aws"
)

type Route53 struct {
	svc          *route53.Route53
	hostedZoneId *string
}

func New(hostedZoneId *string) *Route53 {
	result := Route53{hostedZoneId: hostedZoneId}

	sess, conf := awshelper.GetAwsSession()
	result.svc = route53.New(sess, conf)

	return &result
}

func (r *Route53) CreateChallenge(path string, challenge string) {
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(path),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(`"` + challenge + `"`),
							},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("TXT"),
					},
				},
			},
			Comment: aws.String("ACME challenge"),
		},
		HostedZoneId: r.hostedZoneId,
	}

	result, err := r.svc.ChangeResourceRecordSets(input)
	if err != nil {
		printError(err)
		return
	}

	r.svc.WaitUntilResourceRecordSetsChanged(&route53.GetChangeInput{Id: result.ChangeInfo.Id})
}

func (r *Route53) RemoveChallenge(path string, challenge string) {
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(path),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(`"` + challenge + `"`),
							},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("TXT"),
					},
				},
			},
			Comment: aws.String("ACME challenge"),
		},
		HostedZoneId: r.hostedZoneId,
	}

	result, err := r.svc.ChangeResourceRecordSets(input)
	if err != nil {
		printError(err)
		return
	}

	r.svc.WaitUntilResourceRecordSetsChanged(&route53.GetChangeInput{Id: result.ChangeInfo.Id})
}

func printError(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case route53.ErrCodeNoSuchHostedZone:
			log.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
		case route53.ErrCodeNoSuchHealthCheck:
			log.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
		case route53.ErrCodeInvalidChangeBatch:
			log.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
		case route53.ErrCodeInvalidInput:
			log.Println(route53.ErrCodeInvalidInput, aerr.Error())
		case route53.ErrCodePriorRequestNotComplete:
			log.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
		default:
			log.Println(aerr.Error())
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Println(err.Error())
	}
}
