package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/ockendenjo/strava/pkg/handler"
	"github.com/ockendenjo/strava/pkg/strava"
	"net/http"
	"time"
)

type apiHandler = handler.Handler[events.APIGatewayV2HTTPRequest, events.APIGatewayV2HTTPResponse]

func main() {
	handler.BuildAndStart(func(awsConfig aws.Config) apiHandler {
		ssmClient := ssm.NewFromConfig(awsConfig)

		httpClient := &http.Client{
			Timeout:   3 * time.Second,
			Transport: xray.RoundTripper(http.DefaultTransport),
		}
		stravaClient := strava.NewClient(ssmClient, httpClient)

		return getHandler(stravaClient)
	})
}

func getHandler(client *strava.Client) apiHandler {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		code, found := event.QueryStringParameters["code"]
		if !found {
			return getResponse(http.StatusUnauthorized, "No 'code' parameter in query string parameters"), nil
		}

		err := client.Authorize(ctx, code)
		if err != nil {
			return getResponse(http.StatusInternalServerError, "Something went wrong"), nil
		}
		return getResponse(http.StatusOK, "Authorized"), nil
	}
}

func getResponse(code int, message string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Body:       message,
	}
}
