package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	apigw "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	apigw_integrations "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	lambda "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
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

	connectHandler := lambda.NewFunction(stack, jsii.String("connect"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("connect"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/out", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
	})

	disconnectHandler := lambda.NewFunction(stack, jsii.String("disconnect"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("disconnect"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/out", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
	})

	defaultHandler := lambda.NewFunction(stack, jsii.String("default"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("default"),
		Code:         lambda.Code_FromAsset(jsii.Sprintf("%v/lambdas/websocket/out", dir), nil),
		Architecture: lambda.Architecture_ARM_64(),
	})

	webSocketApi := apigw.NewWebSocketApi(stack, jsii.String("DyscordWSAPI"), &apigw.WebSocketApiProps{
		ConnectRouteOptions: &apigw.WebSocketRouteOptions{
			Integration: apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("ConnectIntegration"), connectHandler, nil),
		},
		DisconnectRouteOptions: &apigw.WebSocketRouteOptions{
			Integration: apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("DisconnectIntegration"), disconnectHandler, nil),
		},
		DefaultRouteOptions: &apigw.WebSocketRouteOptions{
			Integration: apigw_integrations.NewWebSocketLambdaIntegration(jsii.String("DefaultIntegration"), defaultHandler, nil),
		},
	})

	apigw.NewWebSocketStage(stack, jsii.String("DyscordWS"), &apigw.WebSocketStageProps{
		WebSocketApi: webSocketApi,
		StageName:    jsii.String("dev"),
		Description:  jsii.String("My Stage"),
		AutoDeploy:   jsii.Bool(true),
	})

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
