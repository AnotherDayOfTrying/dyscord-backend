package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"

	"dyscord-backend/lambdas/services"
)

var (
	api services.APIGatewayManagementClient
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT"))
	cfg.Region = "us-east-2"

	if err != nil {
		fmt.Println("Error loading config")
	}
	api = services.APIGatewayManagementClient{
		Client: apigatewaymanagementapi.NewFromConfig(cfg),
	}
}

// UnmarshalStreamImage converts events.DynamoDBAttributeValue to struct
func UnmarshalStreamImage(attribute map[string]events.DynamoDBAttributeValue) (map[string]interface{}, error) {
	baseAttrMap := make(map[string]interface{})
	for k, v := range attribute {
		baseAttrMap[k] = extractVal(v)
	}
	return baseAttrMap, nil
}

func extractVal(v events.DynamoDBAttributeValue) interface{} {
	var val interface{}
	switch v.DataType() {
	case events.DataTypeString:
		val = v.String()
	case events.DataTypeNumber:
		val, _ = v.Float()
	case events.DataTypeBinary:
		val = v.Binary()
	case events.DataTypeBoolean:
		val = v.Boolean()
	case events.DataTypeNull:
		val = nil
	case events.DataTypeList:
		list := []interface{}{}
		for _, item := range v.List() {
			list = append(list, extractVal(item))
		}
		val = list
	case events.DataTypeMap:
		mapAttr := map[string]interface{}{}
		for k, v := range v.Map() {
			mapAttr[k] = extractVal(v)
		}
		val = mapAttr
	case events.DataTypeBinarySet:
		set := [][]byte{}
		set = append(set, v.BinarySet()...)
		val = set
	case events.DataTypeNumberSet:
		set := []string{}
		set = append(set, v.NumberSet()...)
		val = set
	case events.DataTypeStringSet:
		set := []string{}
		set = append(set, v.StringSet()...)
		val = set
	}
	return val
}

func handler(ctx context.Context, request events.DynamoDBEvent) error {
	var call services.Call
	for _, record := range request.Records {
		if record.EventName == "MODIFY" && record.Change.StreamViewType == "NEW_IMAGE" {
			newImage, err := UnmarshalStreamImage(record.Change.NewImage)
			if err != nil {
				log.Println(err.Error())
			}
			for k, v := range newImage {
				log.Println(k, v)
			}
			newerImage, err := attributevalue.MarshalMap(newImage)
			if err != nil {
				log.Println(err.Error())
			}

			attributevalue.UnmarshalMap(newerImage, &call)

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

			value, err := json.Marshal(map[string]any{
				"action": "update",
				"data":   values,
			})

			if err != nil {
				log.Println("Could not marshal connection sdps")
			}
			api.PostToConnections(ctx, connectionIds, value)
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
