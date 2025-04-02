package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	dyscordconfig "dyscord-backend/config"
	dynamodbclient "dyscord-backend/lambdas/dynamodb"
)

type Request struct {
	dynamodbclient.SDP
	CallId       string `dynamodbav:"call_id" json:"call_id"`
	ConnectionId string `dynamodbav:"connection_id" json:"connection_id"`
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

	fmt.Printf("%v\n", request)
	fmt.Printf("%v\n", request.Body)

	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Could not parse body"}, nil
	}

	fmt.Printf("%v\n", requestBody)

	response, err := db.GetCall(ctx, requestBody.CallId)

	fmt.Printf("%v\n", response)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	_, err = db.JoinCall(ctx, response, requestBody.ConnectionId, requestBody.SDP)

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
