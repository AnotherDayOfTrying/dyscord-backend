package main

import (
	"encoding/json"
	"os"

	dyscordconfig "dyscord-backend/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	apigw "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	apigw_integrations "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	dynamodb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	lambda "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DyscordBackendStackProps struct {
	awscdk.StackProps
}

func NewDyscordBackendStack(scope constructs.Construct, id string, props *DyscordBackendStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	dir, _ := os.Getwd()

	database := dynamodb.NewTable(stack, jsii.String("DyscordDB"), &dynamodb.TableProps{
		TableName: jsii.String(dyscordconfig.TABLENAME),
		PartitionKey: &dynamodb.Attribute{
			Name: jsii.String("call_id"),
			Type: dynamodb.AttributeType_STRING,
		},
		BillingMode:         dynamodb.BillingMode_PAY_PER_REQUEST,
		TimeToLiveAttribute: jsii.String("ttl"),
		Stream:              dynamodb.StreamViewType_NEW_IMAGE,
	})

	updateHandler := lambda.NewFunction(stack, jsii.String("update"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/update", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Events: &[]lambda.IEventSource{awslambdaeventsources.NewDynamoEventSource(database, &awslambdaeventsources.DynamoEventSourceProps{
			StartingPosition: lambda.StartingPosition_LATEST,
		})},
		Environment: &map[string]*string{
			"AWS_ENDPOINT": aws.String(os.Getenv("AWS_ENDPOINT")),
		},
	})

	connectHandler := lambda.NewFunction(stack, jsii.String("connect"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/connect", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	disconnectHandler := lambda.NewFunction(stack, jsii.String("disconnect"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/disconnect", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	defaultHandler := lambda.NewFunction(stack, jsii.String("default"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/default", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	createCallHandler := lambda.NewFunction(stack, jsii.String("createcall"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/createCall", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	joinCallHandler := lambda.NewFunction(stack, jsii.String("joinCall"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/joinCall", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	leaveCallHandler := lambda.NewFunction(stack, jsii.String("leaveCall"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/leaveCall", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	sendMessageHandler := lambda.NewFunction(stack, jsii.String("sendMessage"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/sendMessage", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	connectRequestTemplate, _ := json.Marshal(map[string]interface{}{
		"statusCode":   200,
		"connectionId": "$context.connectionId",
	})

	webSocketApi := apigw.NewWebSocketApi(stack, jsii.String("DyscordWSAPI"), &apigw.WebSocketApiProps{
		ConnectRouteOptions: &apigw.WebSocketRouteOptions{
			Integration: apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("ConnectionIntegration"), connectHandler, nil),
		},
		DisconnectRouteOptions: &apigw.WebSocketRouteOptions{
			Integration: apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("DisconnectIntegration"), disconnectHandler, nil),
		},
		DefaultRouteOptions: &apigw.WebSocketRouteOptions{
			Integration: apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("DefaultIntegration"), defaultHandler, nil),
		},
	})

	webSocketApi.AddRoute(jsii.String("connectionId"), &apigw.WebSocketRouteOptions{
		Integration: apigw_integrations.NewWebSocketMockIntegration(jsii.String("connectionId"), &apigw_integrations.WebSocketMockIntegrationProps{
			RequestTemplates: &map[string]*string{
				"\\$default": jsii.String(string(connectRequestTemplate)),
			},
			TemplateSelectionExpression: jsii.String("\\$default"),
		}),
		ReturnResponse: jsii.Bool(true),
	})

	webSocketApi.AddRoute(jsii.String("createCall"), &apigw.WebSocketRouteOptions{
		Integration:    apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("CreateCall"), createCallHandler, nil),
		ReturnResponse: jsii.Bool(true),
	})

	webSocketApi.AddRoute(jsii.String("joinCall"), &apigw.WebSocketRouteOptions{
		Integration:    apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("JoinCall"), joinCallHandler, nil),
		ReturnResponse: jsii.Bool(true),
	})

	webSocketApi.AddRoute(jsii.String("sendMessage"), &apigw.WebSocketRouteOptions{
		Integration:    apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("SendMessage"), sendMessageHandler, nil),
		ReturnResponse: jsii.Bool(true),
	})

	webSocketApi.AddRoute(jsii.String("leaveCall"), &apigw.WebSocketRouteOptions{
		Integration:    apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("LeaveCall"), leaveCallHandler, nil),
		ReturnResponse: jsii.Bool(true),
	})

	gateway := apigw.NewWebSocketStage(stack, jsii.String("DyscordWS"), &apigw.WebSocketStageProps{
		WebSocketApi: webSocketApi,
		StageName:    jsii.String("dev"),
		Description:  jsii.String("My Stage"),
		AutoDeploy:   jsii.Bool(true),
	})

	// grant permissions to lambda handlers

	functions := []lambda.Function{
		connectHandler,
		disconnectHandler,
		defaultHandler,
		joinCallHandler,
		createCallHandler,
		sendMessageHandler,
		leaveCallHandler,
	}

	for _, f := range functions {
		database.GrantReadWriteData(f)
	}

	for _, f := range functions {
		gateway.GrantManagementApiAccess(f)
	}

	database.GrantStreamRead(updateHandler)
	gateway.GrantManagementApiAccess(updateHandler)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewDyscordBackendStack(app, "DyscordBackendStack", &DyscordBackendStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
