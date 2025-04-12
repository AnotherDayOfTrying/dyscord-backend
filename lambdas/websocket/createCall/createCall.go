package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	dyscordconfig "dyscord-backend/config"
	"dyscord-backend/lambdas/services"
)

var (
	db services.CallDatabase
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Error loading config")
	}
	db = services.CallDatabase{
		Client:    dynamodb.NewFromConfig(cfg),
		TableName: dyscordconfig.TABLENAME,
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestBody services.SDP
	log.Printf("%v\n", request)
	log.Printf("%v\n", ctx)
	log.Printf("%v\n", request.Body)
	log.Printf("%v\n", []byte(request.Body))
	log.Println("Unmarshalling")
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Could not parse body"}, nil
	}

	log.Printf("%v", requestBody)

	hasher := sha1.New()
	hasher.Write([]byte(time.Now().GoString()))
	sha1_hash := hex.EncodeToString(hasher.Sum(nil))[:6]
	for { // loop until no collision
		data, err := db.GetCall(ctx, sha1_hash)
		log.Println(err, sha1_hash)
		if data.CallId != sha1_hash { // we did not find it
			break
		}
		hasher.Reset()
		hasher.Write([]byte(time.Now().GoString()))
		sha1_hash = hex.EncodeToString(hasher.Sum(nil))[:6]
	}

	log.Println("Created hash: ", sha1_hash)
	err := db.CreateCall(ctx, services.Call{
		CallId:         sha1_hash,
		ConnectionIds:  []string{}, //make connectionids
		ConnectionSdps: []services.SDP{requestBody},
		TTL:            time.Now().Add(time.Hour * 24).Unix(),
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	responseBody, err := json.Marshal(map[string]string{
		"message": "Hello World!",
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
