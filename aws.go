package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetCreds is just an example on how to get environment credentials
func getCreds() (credentials.Value, error) {
	creds := credentials.NewEnvCredentials()
	credValue, err := creds.Get()
	return credValue, err
}

func s3Example() {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: "reticle",
	})
	if err != nil {
		log.Fatal(err)
	}

	svc := s3.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	log.Println(svc)
}
