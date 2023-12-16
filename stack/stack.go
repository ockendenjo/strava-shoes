package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdanodejs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"strings"
)

type StackProps struct {
	awscdk.StackProps
}

func NewStack(scope constructs.Construct, id string, props *StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	clientId := awscdk.NewCfnParameter(stack, jsii.String("clientId"), &awscdk.CfnParameterProps{
		Type:        jsii.String("Number"),
		Description: jsii.String("Strava app client ID"),
		Default:     "0",
	})

	clientSecret := awscdk.NewCfnParameter(stack, jsii.String("clientSecret"), &awscdk.CfnParameterProps{
		Type:        jsii.String("String"),
		Description: jsii.String("Strava app client secret"),
		Default:     "",
		NoEcho:      jsii.Bool(true),
	})

	gearIds := awscdk.NewCfnParameter(stack, jsii.String("GearIds"), &awscdk.CfnParameterProps{
		Type:        jsii.String("String"),
		Description: jsii.String("Stringified JSON of gear IDs to warn about"),
		Default:     `["g9558316"]`,
	})

	role := awsiam.NewRole(stack, jsii.String("LambdaRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})
	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:     jsii.String("AllowSSM"),
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("ssm:PutParameter", "ssm:GetParameters", "ssm:GetParametersByPath", "ssm:DescribeParameters"),
		Resources: jsii.Strings(
			fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/strava*", *stack.Region(), *stack.Account()),
		),
	}))

	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("HttpApi"), &awsapigatewayv2.HttpApiProps{
		ApiName:            jsii.String("StravaShoesApi"),
		CreateDefaultStage: jsii.Bool(true),
	})

	topic := awssns.NewTopic(stack, jsii.String("Topic"), nil)
	topic.GrantPublish(role)

	checkLambda := awslambdanodejs.NewNodejsFunction(stack, jsii.String("ShoesCheckLambda"), &awslambdanodejs.NodejsFunctionProps{
		Role:       role,
		Runtime:    awslambda.Runtime_NODEJS_18_X(),
		Entry:      jsii.String("lib/check-lambda.function.ts"),
		MemorySize: jsii.Number(128),
		Timeout:    awscdk.Duration_Seconds(jsii.Number(5)),
		Environment: &map[string]*string{
			"TOPIC_ARN": topic.TopicArn(),
			"GEAR_IDS":  gearIds.ValueAsString(),
		},
	})

	awsevents.NewRule(stack, jsii.String("CheckRule"), &awsevents.RuleProps{
		Schedule: awsevents.Schedule_Cron(&awsevents.CronOptions{
			Minute: jsii.String("*"),
			Hour:   jsii.String("18"),
		}),
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewLambdaFunction(checkLambda, &awseventstargets.LambdaFunctionProps{
				MaxEventAge:   awscdk.Duration_Minutes(jsii.Number(1)),
				RetryAttempts: jsii.Number(1),
			}),
		},
	})

	authLambda := awslambdanodejs.NewNodejsFunction(stack, jsii.String("AuthLambda"), &awslambdanodejs.NodejsFunctionProps{
		Role:       role,
		Runtime:    awslambda.Runtime_NODEJS_18_X(),
		Entry:      jsii.String("lib/auth-lambda.function.ts"),
		MemorySize: jsii.Number(128),
		Timeout:    awscdk.Duration_Seconds(jsii.Number(5)),
	})
	authIntegration := awsapigatewayv2integrations.NewHttpLambdaIntegration(jsii.String("AuthIntegration"), authLambda, nil)
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/auth"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: authIntegration,
	})

	awsssm.NewStringParameter(stack, jsii.String("ClientIdParameter"), &awsssm.StringParameterProps{
		ParameterName: jsii.String("/strava/clientId"),
		Tier:          awsssm.ParameterTier_STANDARD,
		StringValue:   clientId.ValueAsString(),
	})

	awsssm.NewStringParameter(stack, jsii.String("ClientSecretParameter"), &awsssm.StringParameterProps{
		ParameterName: jsii.String("/strava/clientSecret"),
		Tier:          awsssm.ParameterTier_STANDARD,
		StringValue:   clientSecret.ValueAsString(),
	})

	awsssm.NewStringParameter(stack, jsii.String("AccessTokenParameter"), &awsssm.StringParameterProps{
		ParameterName: jsii.String("/strava/accessToken"),
		Tier:          awsssm.ParameterTier_STANDARD,
		StringValue:   jsii.String("placeholder"),
	})

	awsssm.NewStringParameter(stack, jsii.String("RefreshTokenParameter"), &awsssm.StringParameterProps{
		ParameterName: jsii.String("/strava/refreshToken"),
		Tier:          awsssm.ParameterTier_STANDARD,
		StringValue:   jsii.String("placeholder"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("AuthCallbackDomain"), &awscdk.CfnOutputProps{
		ExportName: jsii.String("AuthCallbackDomain"),
		Value:      jsii.String(strings.Replace(*httpApi.ApiEndpoint(), "https://", "", -1)),
	})

	awscdk.NewCfnOutput(stack, jsii.String("StravaAuthUrl"), &awscdk.CfnOutputProps{
		ExportName: jsii.String("StravaAuthUrl"),
		Value:      jsii.String(fmt.Sprintf("https://www.strava.com/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%sauth&scope=activity:read", *clientId.ValueAsString(), *httpApi.Url())),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewStack(app, "StravaShoesStack", &StackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
