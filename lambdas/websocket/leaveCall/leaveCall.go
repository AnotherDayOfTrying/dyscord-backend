package main

import (
	"context"
	dyscordconfig "dyscord-backend/config"
	"dyscord-backend/lambdas/services"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Request struct {
	CallId       string `dynamodbav:"call_id" json:"call_id"`
	ConnectionId string `dynamodbav:"connection_id" json:"connection_id"`
}

var (
	db services.CallDatabase
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Error loading config")
	}
	db = services.CallDatabase{
		Client:    dynamodb.NewFromConfig(cfg),
		TableName: dyscordconfig.TABLENAME,
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestBody Request

	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	call, err := db.GetCall(ctx, requestBody.CallId)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	_, err = db.LeaveCall(ctx, call, requestBody.ConnectionId)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	responseBody, err := json.Marshal(map[string]string{
		"action": "leaveCall",
		"data":   "Successfully Left Call",
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
