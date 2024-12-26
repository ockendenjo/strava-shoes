package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/ockendenjo/strava/pkg/bagging"
	"github.com/ockendenjo/strava/pkg/strava"
)

func buildCheckActivityFunc(gearIds []string, client bagging.Client, ebClient *eventbridge.Client) checkActivityFn {
	sportTypes := []string{"Run", "Hike", "Walk", "Ride"}
	isGearOk := getCheckFn(sportTypes, gearIds)

	return func(ctx context.Context, activity *strava.Activity, ch chan checkActivityResult) {
		gearOk := isGearOk(activity)
		result := checkActivityResult{
			gearOk:   gearOk,
			activity: activity,
		}

		checked, err := client.HasId(ctx, activity.ID)
		if err != nil {
			result.err = fmt.Errorf("error checking ID %d in bagging DB: %w", activity.ID, err)
			ch <- result
			return
		}

		if checked {
			ch <- result
			return
		}

		res, err := ebClient.PutEvents(ctx, &eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{{
				Source:     aws.String("io.ockenden.strava"),
				DetailType: aws.String("StravaActivityBaggingCheck"),
				Detail:     aws.String(fmt.Sprintf(`{"id": %d}`, activity.ID)),
			}},
		})
		if err != nil {
			result.err = fmt.Errorf("error sending EventBridge event for ID %d: %w", activity.ID, err)
			ch <- result
			return
		}
		if res.FailedEntryCount > 0 {
			result.err = fmt.Errorf("error sending EventBridge event for ID %d: failed", activity.ID)
			ch <- result
			return
		}

		err = client.PutId(ctx, activity.ID)
		if err != nil {
			result.err = fmt.Errorf("error putting ID %d in bagging DB: %w", activity.ID, err)
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
