package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"os/signal"
	"github.com/chriscow/cloud-scanner-go/scan"
	"github.com/chriscow/cloud-scanner-go/util"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/joho/godotenv"
	"github.com/nsqio/go-nsq"
)

var (
	db badger.DB
)

const (
	myChannel = "persist"
)

type resultHandler struct{}

func (h resultHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	res := scan.Result{}
	if err := json.Unmarshal(msg.Body, &res); err != nil {
		return err
	}

	err := db.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(res.Slug), msg.Body)
	})

	if err != nil {
		log.Println("error getting badger transaction", err)
	}

	return err
}

type sessionHandler struct{}

func (h sessionHandler) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	s := scan.Session{}
	if err := json.Unmarshal(msg.Body, &s); err != nil {
		return err
	}

	return nil
}

func checkEnv() {
	godotenv.Load()

}

func main() {
	checkEnv()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	appData := path.Join(os.Getenv("APP_DATA"), "badger")

	db, err := badger.Open(badger.DefaultOptions(appData))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rh := resultHandler{}
	go util.StartConsumer(ctx, scan.ResultTopic, myChannel, rh)

	<-sigChan
	cancel()
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
