package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type LambdaConstruct struct {
	Construct constructs.Construct
	LambdaFn  awslambda.Function
}

func (lc *LambdaConstruct) Node() constructs.Node {
	return lc.Construct.Node()
}

func (l *LambdaConstruct) RunAtFixedRate(ruleName string, schedule awsevents.Schedule, input awsevents.RuleTargetInput) *Rule {
	rule := awsevents.NewRule(l.Construct, jsii.String(ruleName), &awsevents.RuleProps{
		Schedule: schedule,
		RuleName: jsii.String(ruleName),
	})

	rule.AddTarget(awseventstargets.NewLambdaFunction(l.LambdaFn, &awseventstargets.LambdaFunctionProps{
		MaxEventAge:   awscdk.Duration_Minutes(jsii.Number(2)),
		RetryAttempts: jsii.Number(1),
		Event:         input,
	}))
	return &Rule{rule}
}

type Rule struct {
	rule awsevents.Rule
}

func (r *Rule) AddCondition(condition awscdk.CfnCondition) {
	for _, construct := range *r.rule.Node().Children() {
		switch construct.(type) {
		case awscdk.CfnResource:
			(construct.(awscdk.CfnResource)).CfnOptions().SetCondition(condition)
		}
	}
}
