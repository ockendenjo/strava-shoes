package handler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type SQSRecordProcessor func(ctx context.Context, record events.SQSMessage) error

type SQSHandler = Handler[events.SQSEvent, events.SQSEventResponse]

// GetSQSHandler returns a lambda handler that will process each SQS message in parallel using the provided processRecord function
func GetSQSHandler(processRecord SQSRecordProcessor) Handler[events.SQSEvent, events.SQSEventResponse] {

	process := func(ctx context.Context, record events.SQSMessage, successChannel chan bool) {
		err := processRecord(ctx, record)
		if err != nil {
			logger := GetLogger(ctx)
			logger.Error("sqs messaging processing failed", "error", err.Error(), "body", record.Body)
			successChannel <- false
			return
		}
		successChannel <- true
	}

	return func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {

		deadline, hasDeadline := ctx.Deadline()
		if !hasDeadline {
			return events.SQSEventResponse{}, errors.New("context must have a deadline set")
		}
		deadline = deadline.Add(-500 * time.Millisecond)
		subCtx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		//Process each SQS message in its own go routine
		routines := []*routineData{}
		for _, record := range event.Records {
			c := make(chan bool)
			data := routineData{
				SuccessChannel: c,
				Record:         record,
				TimeoutTimer:   time.NewTimer(time.Until(deadline)),
			}
			routines = append(routines, &data)
			go process(subCtx, record, c)
		}

		//For each go routine, start another routine to wait for the result or the timeout
		wg := sync.WaitGroup{}
		for _, routine := range routines {
			wg.Add(1)
			go asyncWaitForResult(ctx, routine, &wg)
		}

		//Collect the failures
		wg.Wait()
		failures := []events.SQSBatchItemFailure{}
		for _, r := range routines {
			if r.failed || r.timedOut {
				failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: r.Record.ReceiptHandle})
			}
		}

		return events.SQSEventResponse{BatchItemFailures: failures}, nil
	}
}

func asyncWaitForResult(ctx context.Context, routine *routineData, wg *sync.WaitGroup) {
	select {
	case success := <-routine.SuccessChannel:
		routine.TimeoutTimer.Stop()
		if !success {
			routine.failed = true
		}
		wg.Done()
	case <-routine.TimeoutTimer.C:
		GetLogger(ctx).Error("sqs message processing timed-out", "body", routine.Record.Body)
		routine.timedOut = true
		wg.Done()
	}
}

type routineData struct {
	SuccessChannel chan bool
	Record         events.SQSMessage
	//Need a timer for each goroutine because the channel only receives one value
	TimeoutTimer *time.Timer
	failed       bool
	timedOut     bool
}
