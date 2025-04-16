package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

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

func handler(ctx context.Context, request events.DynamoDBEvent) error {
	var call services.Call
	log.Println(request)
	values := reflect.ValueOf(request)
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		fmt.Println(types.Field(i).Index[0], types.Field(i).Name, values.Field(i))
	}
	for _, record := range request.Records {
		log.Println(record)
		values := reflect.ValueOf(record)
		types := values.Type()
		for i := 0; i < values.NumField(); i++ {
			fmt.Println(types.Field(i).Index[0], types.Field(i).Name, values.Field(i))
		}
		if record.EventName == "MODIFY" && record.Change.StreamViewType == "NEW_IMAGE" {
			if image, ok := any(record.Change.NewImage).(map[string]dynamodbtypes.AttributeValue); ok {
				err := attributevalue.UnmarshalMap(image, &call)

				if err != nil {
					log.Println("Could not unmarshal map")
				}
				connectionIds := make([]string, len(call.ConnectionSdps))
				for index, sdp := range call.ConnectionSdps {
					connectionIds[index] = sdp.ConnectionId
				}

				value, err := json.Marshal(call.ConnectionSdps)

				if err != nil {
					log.Println("Could not marshal connection sdps")
				}

				api.PostToConnections(ctx, connectionIds, value)
			}
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
