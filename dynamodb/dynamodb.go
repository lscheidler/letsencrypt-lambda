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

package dynamodb

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	//"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/lscheidler/letsencrypt-lambda/account"
	//"github.com/lscheidler/letsencrypt-lambda/crypto"
	awshelper "github.com/lscheidler/letsencrypt-lambda/helper/aws"
)

type DynamoDB struct {
	svc       *dynamodb.DynamoDB
	tableName *string
}

func New(tableName *string) *DynamoDB {
	d := &DynamoDB{tableName: tableName}
	if d.tableName == nil {
		defaultTableName := "LetsencryptCA"
		d.tableName = &defaultTableName
	}
	d.initDB()
	return d
}

func (d *DynamoDB) CreateTable() error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Email"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Email"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: d.tableName,
	}

	_, err := d.svc.CreateTable(input)
	return err
}

func (d *DynamoDB) CreateOrLoadAccount(account *account.Account) error {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: account.Email,
			},
		},
		TableName: d.tableName,
	}

	if result, err := d.svc.GetItem(input); err != nil {
		log.Println("Item not found")
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				// Table doesn't exist
				log.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
				if err = d.CreateTable(); err != nil {
					log.Println(err)
					return err
				}
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return account.Create()
	} else if result.Item == nil {
		log.Println("Item not found")
		// item doesn't exist
		return account.Create()
	} else {
		log.Println("Item found")
		// item exists
		return d.loadAccount(account, result)
	}
}

func (d *DynamoDB) LoadAccount(account *account.Account) error {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: account.Email,
			},
		},
		TableName: d.tableName,
	}

	if result, err := d.svc.GetItem(input); err != nil {
		log.Println("Item not found")
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				// Table doesn't exist
				log.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
				if err = d.CreateTable(); err != nil {
					log.Println(err)
					return err
				}
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return err
	} else if result.Item == nil {
		log.Println("Item not found")
		// item doesn't exist
		return fmt.Errorf("Item not found: %s", *account.Email)
	} else {
		log.Println("Item found")
		// item exists
		return d.loadAccount(account, result)
	}
}

func (d *DynamoDB) loadAccount(acc *account.Account, result *dynamodb.GetItemOutput) error {
	log.Println("loadAccount")
	accountcrypt := account.AccountCrypt(*acc)
	if err := json.Unmarshal([]byte(*result.Item["Data"].S), &accountcrypt); err != nil {
		log.Println("loadAccount: Unmarshal error")
		return err
	}
	*acc = account.Account(accountcrypt)
	return nil
}

func (d *DynamoDB) Update(acc *account.Account) error {
	if acc.Changed {
		accountcrypt := account.AccountCrypt(*acc)
		jsonCipher, err := json.Marshal(&accountcrypt)
		if err != nil {
			return err
		}

		input := &dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#D": aws.String("Data"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":d": {
					S: aws.String(string(jsonCipher)),
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				"Email": {
					S: acc.Email,
				},
			},
			ReturnValues:     aws.String("ALL_NEW"),
			TableName:        d.tableName,
			UpdateExpression: aws.String("SET #D = :d"),
		}

		_, err = d.svc.UpdateItem(input)
		return err
	}
	return nil
}

func (d *DynamoDB) initDB() {
	log.Println("Initialize database connection")
	sess, conf := awshelper.GetAwsSession()
	d.svc = dynamodb.New(sess, conf)
}
