package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/ockendenjo/handler"
)

type H = handler.Handler[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse]

func main() {
	handler.BuildAndStart(func(awsConfig aws.Config) H {
		ebClient := eventbridge.NewFromConfig(awsConfig)

		h := &lambdaHandler{ebClient: ebClient}
		return h.handle
	})
}

type lambdaHandler struct {
	ebClient *eventbridge.Client
}

func (h *lambdaHandler) handle(ctx *handler.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := ctx.GetLogger()
	logger.AddParam("event", event.Body).Info("Received event")

	entry := types.PutEventsRequestEntry{
		Detail:     aws.String(event.Body),
		DetailType: aws.String("StravaEvent"),
		Source:     aws.String("io.ockenden.strava"),
	}
	_, err := h.ebClient.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{entry},
	})

	if err != nil {
		logger.Error("Failed to send event to bus")
	} else {
		logger.AddParam("published", entry).Info("Sent event to bus")
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: ""}, nil
}
