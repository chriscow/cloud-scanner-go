package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/joho/godotenv"
)

func checkEnv() {
	godotenv.Load()

}

func main() {
	checkEnv()

	// ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// handler := &scanRadiusHandler{}

	// go util.StartConsumer(ctx, "scan-session", "scan-radius-scanner", handler)

	select {
	case <-sigChan:
		// cancel()
	default:
	}
}

func exampleMarshal() {
	type Record struct {
		ID   string
		URLs []string
	}

	mySession := session.Must(session.NewSession())

	// Create a DynamoDB client with additional configuration
	svc := dynamodb.New(mySession, aws.NewConfig().WithRegion("us-west-2"))

	r := Record{
		ID: "ABC123",
		URLs: []string{
			"https://example.com/first/link",
			"https://example.com/second/url",
		},
	}
	av, err := dynamodbattribute.MarshalMap(r)
	if err != nil {
		panic(fmt.Sprintf("failed to DynamoDB marshal Record, %v", err))
	}

	myTableName := "example"

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(myTableName),
		Item:      av,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to put Record to DynamoDB, %v", err))
	}
}

func exampleUnmarshal() {
	type Record struct {
		ID   string
		URLs []string
	}

	mySession := session.Must(session.NewSession())

	// Create a DynamoDB client with additional configuration
	svc := dynamodb.New(mySession, aws.NewConfig().WithRegion("us-west-2"))

	var records []Record
	myTableName := "example"

	// Use the ScanPages method to perform the scan with pagination. Use
	// just Scan method to make the API call without pagination.
	err := svc.ScanPages(&dynamodb.ScanInput{
		TableName: aws.String(myTableName),
	}, func(page *dynamodb.ScanOutput, last bool) bool {
		recs := []Record{}

		err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &recs)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
		}

		records = append(records, recs...)

		return true // keep paging
	})

	if err != nil {
		log.Fatal(err)
	}
}
