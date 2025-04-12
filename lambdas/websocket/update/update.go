package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"dyscord-backend/lambdas/services"
)

var (
	api services.APIGatewayManagementClient
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Error loading config")
	}
	api = services.APIGatewayManagementClient{
		Client: apigatewaymanagementapi.NewFromConfig(cfg),
	}
}

func handler(ctx context.Context, request events.DynamoDBStreamRecord) error {
	var call services.Call
	if request, ok := any(request.NewImage).(map[string]types.AttributeValue); ok {
		attributevalue.UnmarshalMap(request, &call)
		value, err := json.Marshal(call.ConnectionSdps)
		if err != nil {
			return err
		}
		api.PostToConnections(ctx, call.ConnectionIds, value)
	} else {
		return errors.New("could not unmarshall request")
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
