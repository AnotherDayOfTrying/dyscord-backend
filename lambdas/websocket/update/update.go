package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"

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

func handler(ctx context.Context, request json.RawMessage) error {
	// var call services.Call
	log.Println(request)
	log.Println(string(request))
	// if request, ok := any(request.NewImage).(map[string]types.AttributeValue); ok {
	// 	attributevalue.UnmarshalMap(request, &call)
	// 	value, err := json.Marshal(call.ConnectionSdps)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	connectionIds := make([]string, len(call.ConnectionSdps))
	// 	for index, sdp := range call.ConnectionSdps {
	// 		connectionIds[index] = sdp.ConnectionId
	// 	}
	// 	api.PostToConnections(ctx, connectionIds, value)
	// } else {
	// 	return errors.New("could not unmarshall request")
	// }
	return nil
}

func main() {
	lambda.Start(handler)
}
