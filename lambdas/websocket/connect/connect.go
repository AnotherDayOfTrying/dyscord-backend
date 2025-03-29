package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIGatewayProxyRequest struct {
	events.APIGatewayProxyRequest
	ConnectionId string `json:"connectionId"`
}

func handler(request APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	fmt.Printf("%v", request)

	responseBody, err := json.Marshal(map[string]string{
		"message":      "Connected!",
		"connectionId": request.ConnectionId,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
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
