package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/aws/jsii-runtime-go"
	"github.com/ockendenjo/strava/pkg/handler"
	"github.com/ockendenjo/strava/pkg/strava"
	"net/http"
	"slices"
	"strings"
	"time"
)

func main() {
	gearIds := mustGetSliceEnv("GEAR_IDS")
	topicArn := handler.MustGetEnv("TOPIC_ARN")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		ssmClient := ssm.NewFromConfig(awsConfig)

		httpClient := &http.Client{
			Timeout:   3 * time.Second,
			Transport: xray.RoundTripper(http.DefaultTransport),
		}
		stravaClient := strava.NewClient(ssmClient, httpClient)

		snsClient := sns.NewFromConfig(awsConfig)

		return getHandler(stravaClient, snsClient, gearIds, topicArn)
	})
}

func getHandler(stravaClient *strava.Client, snsClient *sns.Client, gearIds []string, topicArn string) handler.Handler[any, any] {
	return func(ctx context.Context, event any) (any, error) {
		logger := handler.GetLogger(ctx)

		//Load activities
		activities, err := stravaClient.GetActivities(ctx)
		if err != nil {
			return nil, err
		}

		var messages []string
		sportTypes := []string{"Run", "Hike", "Walk"}
		isGearOk := getCheckFn(sportTypes, gearIds)

		for _, activity := range activities {
			if isGearOk(activity) {
				continue
			}

			logger.Warn("Activity with missing gear", "activity", activity)
			msg := fmt.Sprintf("%s (%s) https://www.strava.com/activities/%d", activity.Name, activity.SportType, activity.ID)
			messages = append(messages, msg)
		}

		if len(messages) < 1 {
			logger.Info("No missing gear")
			return nil, nil
		}
		fullMessage := strings.Join(messages, "\n")

		_, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: jsii.String(topicArn),
			Message:  jsii.String(fullMessage),
			Subject:  jsii.String("Strava activities with missing gear"),
		})
		return nil, err
	}
}

func getCheckFn(sportTypes []string, gearIds []string) func(a strava.Activity) bool {

	return func(a strava.Activity) bool {
		if !slices.Contains(sportTypes, a.SportType) {
			//Sport type is ignored
			return true
		}

		if a.GearID == "" {
			return false
		}

		return !slices.Contains(gearIds, a.GearID)
	}
}

func mustGetSliceEnv(key string) []string {
	v := handler.MustGetEnv(key)
	var a []string
	err := json.Unmarshal([]byte(v), &a)
	if err != nil {
		panic(err)
	}
	return a
}
