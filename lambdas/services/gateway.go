package services

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

type APIGatewayManagementClient struct {
	Client *apigatewaymanagementapi.Client
}

func (c *APIGatewayManagementClient) PostToConnections(ctx context.Context, connectionIds []string, data []byte) {
	for _, connectionId := range connectionIds {
		_, err := c.Client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connectionId,
			Data:         data,
		})
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
