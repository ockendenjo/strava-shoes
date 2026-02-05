package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ockendenjo/handler"
)

type H = handler.Handler[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse]

func main() {
	handler.BuildAndStart(func(awsConfig aws.Config) H {
		return func(ctx *handler.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

			ctx.GetLogger().AddParam("event", event.Body).Info("Received event")
			return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: ""}, nil
		}
	})
}
