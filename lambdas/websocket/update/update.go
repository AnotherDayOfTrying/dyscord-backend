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

func UnmarshalStreamImage(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {

	dbAttrMap := make(map[string]dynamodbtypes.AttributeValue)

	for k, v := range attribute {
		log.Println(k, v)
		var dbAttr dynamodbtypes.AttributeValue

		bytes, marshalErr := v.MarshalJSON()
		log.Println("Bytes:", bytes)
		log.Println("String:", string(bytes))
		if marshalErr != nil {
			return marshalErr
		}

		json.Unmarshal(bytes, &dbAttr)
		log.Println(dbAttr)
		dbAttrMap[k] = dbAttr
		log.Println(dbAttrMap)
	}

	return attributevalue.UnmarshalMap(dbAttrMap, out)
}

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
			err := UnmarshalStreamImage(record.Change.NewImage, &call)

			if err != nil {
				log.Println("Could not unmarshal map")
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
