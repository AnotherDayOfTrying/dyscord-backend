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
		list := make([]interface{}, len(v.List()))
		for _, item := range v.List() {
			list = append(list, extractVal(item))
		}
		val = list
	case events.DataTypeMap:
		mapAttr := make(map[string]interface{}, len(v.Map()))
		for k, v := range v.Map() {
			mapAttr[k] = extractVal(v)
		}
		val = mapAttr
	case events.DataTypeBinarySet:
		set := make([][]byte, len(v.BinarySet()))
		for _, item := range v.BinarySet() {
			set = append(set, item)
		}
		val = set
	case events.DataTypeNumberSet:
		set := make([]string, len(v.NumberSet()))
		for _, item := range v.NumberSet() {
			set = append(set, item)
		}
		val = set
	case events.DataTypeStringSet:
		set := make([]string, len(v.StringSet()))
		for _, item := range v.StringSet() {
			set = append(set, item)
		}
		val = set
	}
	return val
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
			newImage, err := UnmarshalStreamImage(record.Change.NewImage)
			if err != nil {
				log.Println(err.Error())
			}
			newerImage, err := attributevalue.MarshalMap(newImage)
			log.Println(newerImage)
			for k, v := range newerImage {
				log.Println(k, v)
			}
			if err != nil {
				log.Println(err.Error())
			}

			attributevalue.UnmarshalMap(newerImage, &call)

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
