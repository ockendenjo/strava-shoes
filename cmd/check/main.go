package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/aws/jsii-runtime-go"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/strava/pkg/bagging"
	"github.com/ockendenjo/strava/pkg/strava"
)

const maxParallel = 10

type H = handler.Handler[CheckActivitiesEvent, any]

func main() {
	gearIds := mustGetSliceEnv("GEAR_IDS")
	topicArn := handler.MustGetEnv("TOPIC_ARN")
	baggingDb := handler.MustGetEnv("BAGGING_DB")

	handler.BuildAndStart(func(awsConfig aws.Config) H {
		ssmClient := ssm.NewFromConfig(awsConfig)
		ebClient := eventbridge.NewFromConfig(awsConfig)
		baggingClient := bagging.NewClient(dynamodb.NewFromConfig(awsConfig), baggingDb)
		checkActivity := buildCheckActivityFunc(gearIds, baggingClient, ebClient)

		httpClient := &http.Client{
			Timeout:   3 * time.Second,
			Transport: xray.RoundTripper(http.DefaultTransport),
		}
		stravaClient := strava.NewClient(ssmClient, httpClient)

		snsClient := sns.NewFromConfig(awsConfig)

		return getHandler(stravaClient, snsClient, checkActivity, topicArn)
	})
}

func getHandler(stravaClient *strava.Client, snsClient *sns.Client, checkActivity checkActivityFn, topicArn string) H {
	return func(ctx *handler.Context, event CheckActivitiesEvent) (any, error) {
		logger := ctx.GetLogger()

		page := 1
		if event.Page > 1 {
			page = event.Page
		}

		//Load activities
		activities, err := stravaClient.GetActivities(ctx, page)
		if err != nil {
			return nil, err
		}

		var messages []string

		ch := make(chan checkActivityResult, maxParallel)
		remaining := 0
		var parrallelError error

		readChan := func() {
			res := <-ch
			remaining--
			if res.err != nil {
				parrallelError = res.err
			}
			activity := res.activity
			if !res.gearOk {
				logger.Warn("Activity with missing gear", "activity", activity)
				msg := fmt.Sprintf("%s (%s) https://www.strava.com/activities/%d", activity.Name, activity.SportType, activity.ID)
				messages = append(messages, msg)
			}
		}

		for i, activity := range activities {
			go checkActivity(ctx, &activity, ch)
			remaining++

			if i > maxParallel {
				readChan()
			}
		}
		for remaining > 0 {
			readChan()
		}
		if parrallelError != nil {
			return nil, parrallelError
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

func mustGetSliceEnv(key string) []string {
	v := handler.MustGetEnv(key)
	var a []string
	err := json.Unmarshal([]byte(v), &a)
	if err != nil {
		panic(err)
	}
	return a
}

type CheckActivitiesEvent struct {
	Page int `json:"page,omitempty"`
}
