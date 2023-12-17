package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	lambda "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type LambdaBuilder struct {
	scope        constructs.Construct
	id           string
	codePath     string
	functionName string
	env          map[string]*string
	role         awsiam.Role
}

func NewLambda(stack awscdk.Stack, id string, codePath string) *LambdaBuilder {
	construct := constructs.NewConstruct(stack, jsii.String(id))

	role := awsiam.NewRole(construct, jsii.String("serviceRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(fmt.Sprintf("lambda-%s__%s", *stack.Region(), id)),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
	})
	lambdaPermissions := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("logs:*"),
		},
		Resources: &[]*string{
			jsii.String("*"),
		},
	})
	role.AddToPolicy(lambdaPermissions)

	return &LambdaBuilder{
		scope:        construct,
		id:           id,
		codePath:     codePath,
		functionName: "Strava_" + id,
		env:          map[string]*string{},
		role:         role,
	}
}

func (lb *LambdaBuilder) WithTopicPublish(topic awssns.Topic, topicArnEnvKey string) *LambdaBuilder {
	topic.GrantPublish(lb.role)
	lb.env[topicArnEnvKey] = topic.TopicArn()
	return lb
}

func (lb *LambdaBuilder) WithParamsAccess() *LambdaBuilder {
	stack := lb.role.Stack()

	lb.role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:     jsii.String("AllowSSM"),
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("ssm:PutParameter", "ssm:GetParameters", "ssm:GetParametersByPath", "ssm:DescribeParameters"),
		Resources: jsii.Strings(
			fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/strava*", *stack.Region(), *stack.Account()),
		),
	}))

	return lb
}

func (lb *LambdaBuilder) WithDynamoDB(table awsdynamodb.Table, tableEnvKey string) *LambdaBuilder {
	table.GrantReadWriteData(lb.role)
	lb.env[tableEnvKey] = table.TableName()
	return lb
}

func (lb *LambdaBuilder) WithHttpApi(api awsapigateway.RestApi, urlEnvKey string) *LambdaBuilder {
	lb.env[urlEnvKey] = api.Url()
	return lb
}

func (lb *LambdaBuilder) WithEnvVar(envValue string, envKey string) *LambdaBuilder {
	lb.env[envKey] = jsii.String(envValue)
	return lb
}

func (lb *LambdaBuilder) Build() *LambdaConstruct {

	lambdaFn := lambda.NewFunction(lb.scope, jsii.String("function"), &lambda.FunctionProps{
		Runtime:      lambda.Runtime_PROVIDED_AL2023(),
		Architecture: lambda.Architecture_ARM_64(),
		Handler:      jsii.String("function"),
		Role:         lb.role,
		Code:         lambda.Code_FromAsset(jsii.String(lb.codePath), nil),
		FunctionName: jsii.String(lb.functionName),
		Environment:  &lb.env,
		Timeout:      awscdk.Duration_Seconds(jsii.Number(10)),
		MemorySize:   jsii.Number(256),
		Tracing:      lambda.Tracing_ACTIVE,
	})

	return &LambdaConstruct{LambdaFn: lambdaFn, Construct: lb.scope}
}

func DailyAtTime(hour int, minutes int) awsevents.Schedule {
	return awsevents.Schedule_Cron(&awsevents.CronOptions{
		Day:    jsii.String("*"),
		Hour:   jsii.String(fmt.Sprintf("%d", hour)),
		Minute: jsii.String(fmt.Sprintf("%d", minutes)),
	})
}
