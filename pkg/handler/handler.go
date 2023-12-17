package handler

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

const loggerKey = "logger"

func GetLogger(ctx context.Context) *slog.Logger {
	val := ctx.Value(loggerKey)
	if val != nil {
		return val.(*slog.Logger)
	}
	return slog.Default()
}

type Handler[T interface{}, U interface{}] func(ctx context.Context, event T) (U, error)

func WithLogger[T interface{}, U interface{}](handlerFunc Handler[T, U]) Handler[T, U] {
	return func(ctx context.Context, event T) (U, error) {
		// Perform pre-handler tasks here
		newContext := ContextWithLogger(ctx)

		response, err := handlerFunc(newContext, event)
		if err != nil {
			logger := GetLogger(ctx)
			logger.Error("lambda execution failed", "error", err.Error())
		}

		return response, err
	}
}

func ContextWithLogger(ctx context.Context) context.Context {
	traceId := os.Getenv("_X_AMZN_TRACE_ID")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if traceId != "" {
		parts := strings.Split(traceId, ";")
		if len(parts) > 0 {
			logger = logger.With("trace_id", strings.Replace(parts[0], "Root=", "", 1))
		}
	}
	newContext := context.WithValue(ctx, loggerKey, logger)
	return newContext
}

func MustGetEnv(key string) string {
	val := os.Getenv(key)
	if strings.Trim(val, " ") == "" {
		panic(fmt.Errorf("environment variable for '%s' has not been set", key))
	}
	return val
}

// BuildAndStart configures a logger, instruments the handler with OpenTelemetry, instruments the AWS SDK, and then starts the lambda
func BuildAndStart[T interface{}, U interface{}](getHandler func(awsConfig aws.Config) Handler[T, U]) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	//Instrument the AWS SDK - this needs to happen before any service clients (e.g. s3Client) are created
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)

	//Pass the AWS config to the get handler - service clients can be created in this method
	handlerFn := getHandler(cfg)

	lambda.Start(WithLogger(handlerFn))
}
