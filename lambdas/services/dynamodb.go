package services

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Call struct {
	CallId         string `dynamodbav:"call_id" json:"call_id"`
	ConnectionSdps []SDP  `dynamodbav:"connection_sdps" json:"connection_sdps"`
	TTL            int64  `dynamodbav:"ttl" json:"ttl"`
}

type SDP struct {
	ConnectionId               string `dynamodbav:"connection_id" json:"connection_id"`
	Type                       string `dynamodbav:"type" json:"type"`
	SessionDescriptionProtocol string `dynamodbav:"sdp" json:"sdp"`
}

func (call Call) GetKey() map[string]types.AttributeValue {
	callId, err := attributevalue.Marshal(call.CallId)
	if err != nil {
		panic(err)
	}

	return map[string]types.AttributeValue{"call_id": callId}
}

type CallDatabase struct {
	Client    *dynamodb.Client
	TableName string
}

func (db CallDatabase) CreateCall(ctx context.Context, call Call) error {
	var response *dynamodb.PutItemOutput
	var responseValues map[string]map[string]interface{}
	item, err := attributevalue.MarshalMap(call)
	if err != nil {
		panic(err)
	}

	response, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.TableName),
		Item:      item,
	})

	if err != nil {
		log.Printf("Item could not be added, %v", err)
	}

	err = attributevalue.UnmarshalMap(response.Attributes, &responseValues)
	if err != nil {
		log.Printf("Could not unmarshal response, %v", err)
	}
	return err
}

func (db CallDatabase) GetCall(ctx context.Context, callId string) (Call, error) {
	call := Call{CallId: callId}
	response, err := db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:       call.GetKey(),
		TableName: aws.String(db.TableName),
	})
	if err != nil {
		log.Printf("Item could not be got, %v", err)
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &call)
		if err != nil {
			log.Printf("Failed to Unmarshal Item, %v", err)
		}
	}
	return call, err
}

func (db CallDatabase) JoinCall(ctx context.Context, call Call, connectionId string, sdp SDP) (map[string]interface{}, error) {
	var response *dynamodb.UpdateItemOutput
	var responseValues map[string]interface{}

	marshalledSdp, err := attributevalue.MarshalMap(sdp)

	if err != nil {
		log.Println(err.Error())
	}

	update := expression.Set(
		expression.Name("connection_ids"),
		expression.ListAppend(
			expression.IfNotExists(expression.Name("connection_ids"), expression.Value(&types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: connectionId}}})),
			expression.Value(&types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: connectionId}}}),
		),
	).Set(
		expression.Name("connection_sdps"),
		expression.ListAppend(
			expression.IfNotExists(expression.Name("connection_sdps"), expression.Value(&types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberM{Value: marshalledSdp}}})),
			expression.Value(&types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberM{Value: marshalledSdp}}}),
		),
	)
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("Item could not build expression, %v", err)
	} else {
		response, err = db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:                 aws.String(db.TableName),
			Key:                       call.GetKey(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ReturnValues:              types.ReturnValueUpdatedNew,
		})
		if err != nil {
			log.Printf("Item could not be updated, %v", err)
		} else {
			err = attributevalue.UnmarshalMap(response.Attributes, &responseValues)
			if err != nil {
				log.Printf("Unable to unmarshal map, %v", err)
			}
		}
	}

	return responseValues, err
}

func (db CallDatabase) LeaveCall(ctx context.Context, call Call, connectionId string) (map[string]any, error) {
	var responseValues map[string]any

	originalCall, err := db.GetCall(ctx, call.CallId)

	if err != nil {
		log.Println(err.Error())
	}

	found := false
	connectionIndex := -1
	for index, value := range originalCall.ConnectionSdps {
		if connectionId == value.ConnectionId {
			found = true
			connectionIndex = index
			break
		}
	}

	if !found {
		return nil, nil
	}

	update := expression.Remove(
		expression.Name(fmt.Sprintf("connection_ids[%d]", connectionIndex)),
	)

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("Item could not build expression, %v", err)
	} else {
		_, err = db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:                 aws.String(db.TableName),
			Key:                       call.GetKey(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ReturnValues:              types.ReturnValueUpdatedNew,
		})
		if err != nil {
			log.Printf("Item could not be updated, %v", err)
		}
	}
	condition := expression.Name("connection_ids").Size().Equal(expression.Value(0))
	expr, err = expression.NewBuilder().WithCondition(condition).Build()
	if err != nil {
		log.Printf("Item could not build expression, %v", err)
	} else {
		response, err := db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName:                 aws.String(db.TableName),
			Key:                       call.GetKey(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			ConditionExpression:       expr.Condition(),
			ReturnValues:              types.ReturnValueNone,
		})
		if err != nil {
			log.Printf("Item could not be updated, %v", err)
		} else {
			err = attributevalue.UnmarshalMap(response.Attributes, &responseValues)
			if err != nil {
				log.Printf("Unable to unmarshal map, %v", err)
			}
		}
	}

	return responseValues, err
}
