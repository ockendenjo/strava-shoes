package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/strava/services/ps"
)

type H = handler.Handler[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse]

func main() {
	handler.BuildAndStart(func(awsConfig aws.Config) H {

		ssmClient := ssm.NewFromConfig(awsConfig)
		paramsClient := ps.NewParamsClient(ssmClient)
		h := &lambdaHandler{paramsClient: paramsClient}
		return h.handle
	})
}

type lambdaHandler struct {
	paramsClient ps.ParamsClient
}

func (h *lambdaHandler) handle(ctx *handler.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := ctx.GetLogger()

	challenge, found := event.QueryStringParameters["hub.challenge"]
	if !found {
		logger.Error("Challenge not found")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	verifyToken, found := event.QueryStringParameters["hub.verify_token"]
	if !found {
		logger.Error("VerifyToken not found")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	params, err := h.paramsClient.GetParams(ctx)
	if err != nil {
		logger.Error("Failed to get params")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	if params.VerifyToken != verifyToken {
		logger.Error("Verify token does not match")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	resp := ChallengeResponse{Challenge: challenge}
	b, err := json.Marshal(resp)
	if err != nil {
		logger.AddParam("error", err).Error("Error")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	logger.Info("Responding with HTTP 200")
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(b),
	}, nil
}

type ChallengeResponse struct {
	Challenge string `json:"hub.challenge"`
}
