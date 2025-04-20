package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"

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

func handler(ctx context.Context, request dynamodbstreams.GetRecordsOutput) error {
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
		if record.EventName == "MODIFY" && record.Dynamodb.StreamViewType == "NEW_IMAGE" {
			newImage, err := attributevalue.FromDynamoDBStreamsMap(record.Dynamodb.NewImage)
			log.Println(newImage)
			if err != nil {
				log.Println(err.Error())
			}
			err = attributevalue.UnmarshalMap(newImage, &call)

			if err != nil {
				log.Println(err.Error())
			}

			test := reflect.ValueOf(call)
			types := test.Type()
			for i := 0; i < test.NumField(); i++ {
				fmt.Println(types.Field(i).Index[0], types.Field(i).Name, test.Field(i))
			}

			connectionIds := make([]string, len(call.ConnectionSdps))
			values := make([]interface{}, len(call.ConnectionSdps))
			for index, sdp := range call.ConnectionSdps {
				connectionIds[index] = sdp.ConnectionId
				values[index] = struct {
					Type                       string `dynamodbav:"type" json:"type"`
					SessionDescriptionProtocol string `dynamodbav:"sdp" json:"sdp"`
				}{
					Type:                       sdp.Type,
					SessionDescriptionProtocol: sdp.SessionDescriptionProtocol,
				}
			}

			value, err := json.Marshal(values)

			if err != nil {
				log.Println("Could not marshal connection sdps")
			}
			log.Println(connectionIds)
			log.Println(values)
			log.Println(value)
			api.PostToConnections(ctx, connectionIds, value)
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
