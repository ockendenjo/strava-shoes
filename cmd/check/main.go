package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/aws/jsii-runtime-go"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/strava/pkg/bagging"
	"github.com/ockendenjo/strava/pkg/strava"
)

const maxParallel = 10

func main() {
	gearIds := mustGetSliceEnv("GEAR_IDS")
	topicArn := handler.MustGetEnv("TOPIC_ARN")
	baggingDb := handler.MustGetEnv("BAGGING_DB")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		ssmClient := ssm.NewFromConfig(awsConfig)
		baggingClient := bagging.NewClient(dynamodb.NewFromConfig(awsConfig), baggingDb)
		checkActivity := buildCheckActivityFunc(gearIds, baggingClient)

		httpClient := &http.Client{
			Timeout:   3 * time.Second,
			Transport: xray.RoundTripper(http.DefaultTransport),
		}
		stravaClient := strava.NewClient(ssmClient, httpClient)

		snsClient := sns.NewFromConfig(awsConfig)

		return getHandler(stravaClient, snsClient, checkActivity, topicArn)
	})
}

func getHandler(stravaClient *strava.Client, snsClient *sns.Client, checkActivity checkActivityFn, topicArn string) handler.Handler[any, any] {
	return func(ctx context.Context, event any) (any, error) {
		logger := handler.GetLogger(ctx)

		//Load activities
		activities, err := stravaClient.GetActivities(ctx)
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

func buildCheckActivityFunc(gearIds []string, client bagging.Client) checkActivityFn {
	sportTypes := []string{"Run", "Hike", "Walk", "Ride"}
	isGearOk := getCheckFn(sportTypes, gearIds)

	return func(ctx context.Context, activity *strava.Activity, ch chan checkActivityResult) {
		gearOk := isGearOk(activity)
		result := checkActivityResult{
			gearOk: gearOk,
		}

		checked, err := client.HasId(ctx, activity.ID)
		if err != nil {
			result.err = err
			ch <- result
			return
		}

		if checked {
			ch <- result
			return
		}

		//TODO: Send to EventBridge

		err = client.PutId(ctx, activity.ID)
		if err != nil {
			result.err = err
			ch <- result
			return
		}

		ch <- result
	}
}

type checkActivityResult struct {
	gearOk   bool
	err      error
	activity *strava.Activity
}

type checkActivityFn func(ctx context.Context, activity *strava.Activity, ch chan checkActivityResult)

func getCheckFn(sportTypes []string, gearIds []string) func(a *strava.Activity) bool {

	return func(a *strava.Activity) bool {
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
