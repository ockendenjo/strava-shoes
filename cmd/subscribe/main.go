package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/google/uuid"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/strava/pkg/strava"
)

func main() {

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		callbackURL := handler.MustGetEnv("CALLBACK_URL")
		ssmClient := ssm.NewFromConfig(awsConfig)

		httpClient := &http.Client{
			Timeout:   3 * time.Second,
			Transport: xray.RoundTripper(http.DefaultTransport),
		}
		stravaClient := strava.NewClient(ssmClient, httpClient)

		h := &lambdaHandler{
			callbackURL:  callbackURL,
			stravaClient: stravaClient,
		}
		return h.handle
	})

}

type lambdaHandler struct {
	stravaClient *strava.Client
	callbackURL  string
}

func (h *lambdaHandler) handle(ctx *handler.Context, event any) (any, error) {
	logger := ctx.GetLogger()
	verifyToken := strings.ReplaceAll(uuid.NewString(), "-", "")

	err := h.stravaClient.Subscribe(ctx, h.callbackURL, verifyToken)
	if err != nil {
		logger.AddParam("error", err).Error("Error subscribing to events")
		return nil, err
	}

	logger.Info("Subscribed to events")
	return event, nil
}
