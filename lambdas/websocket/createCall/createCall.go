package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	dyscordconfig "dyscord-backend/config"
	dynamodbclient "dyscord-backend/lambdas/dynamodb"
)

// we need to post to connection with connectionId

type Request struct {
	Type                       string `json:"type"`
	SessionDescriptionProtocol string `json:"sdp"`
}

var (
	db dynamodbclient.CallDatabase
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Error loading config")
	}
	db = dynamodbclient.CallDatabase{
		Client:    dynamodb.NewFromConfig(cfg),
		TableName: dyscordconfig.TABLENAME,
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestBody Request
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err == nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Could not parse body"}, nil
	}

	log.Printf("%v", requestBody)

	err := db.CreateCall(ctx, dynamodbclient.Call{
		CallId:                     "111111",   // !!!TODO: CHANGE THIS
		ConnectionIds:              []string{}, //make connectionids
		Type:                       requestBody.Type,
		SessionDescriptionProtocol: requestBody.SessionDescriptionProtocol, // make sdp
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	responseBody, err := json.Marshal(map[string]string{
		"message": "Hello World!",
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Sever Error"}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
