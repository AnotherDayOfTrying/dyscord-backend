package dynamodbclient

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Call struct {
	CallId                     string   `dynamodbav:"call_id"`
	ConnectionIds              []string `dynamodbav:"connection_ids"`
	Type                       string   `dynamodbav:"type"`
	SessionDescriptionProtocol string   `dynamodbav:"sdp"`
	TTL                        int64    `dynamodbav:"ttl"`
}

func (call Call) GetKey() map[string]types.AttributeValue {
	roomId, err := attributevalue.Marshal(call.CallId)
	if err != nil {
		panic(err)
	}

	return map[string]types.AttributeValue{"test": roomId}
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

func (db CallDatabase) JoinCall(ctx context.Context, call Call, connectionId string) (map[string]map[string]interface{}, error) {
	var response *dynamodb.UpdateItemOutput
	var responseValues map[string]map[string]interface{}

	update := expression.Set(
		expression.Name("connection_ids"),
		expression.ListAppend(
			expression.IfNotExists(expression.Name("connection_ids"), expression.Value([]string{})),
			expression.Value(connectionId),
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

func (db CallDatabase) LeaveCall(ctx context.Context, call Call, connectionId string) (map[string]map[string]interface{}, error) {
	var response *dynamodb.UpdateItemOutput
	var responseValues map[string]map[string]interface{}

	update := expression.Delete(
		expression.Name("connection_ids"),
		expression.Value(connectionId),
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
		response, err = db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:                 aws.String(db.TableName),
			Key:                       call.GetKey(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ConditionExpression:       expr.Condition(),
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
