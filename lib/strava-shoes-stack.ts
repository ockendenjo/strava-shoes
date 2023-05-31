import {CfnOutput, CfnParameter, Duration, Stack, StackProps} from "aws-cdk-lib";
import {Construct} from "constructs";
import {NodejsFunction} from "aws-cdk-lib/aws-lambda-nodejs";
import {Runtime} from "aws-cdk-lib/aws-lambda";
import {HttpApi, HttpMethod} from "@aws-cdk/aws-apigatewayv2-alpha";
import {Effect, PolicyStatement, Role, ServicePrincipal} from "aws-cdk-lib/aws-iam";
import {Topic} from "aws-cdk-lib/aws-sns";
import {HttpLambdaIntegration} from "@aws-cdk/aws-apigatewayv2-integrations-alpha";
import {ParameterDataType, ParameterTier, StringParameter} from "aws-cdk-lib/aws-ssm";
import {Rule, Schedule} from "aws-cdk-lib/aws-events";
import {LambdaFunction} from "aws-cdk-lib/aws-events-targets";

export class StravaShoesStack extends Stack {
    constructor(scope: Construct, id: string, props?: StackProps) {
        super(scope, id, props);

        const clientId = new CfnParameter(this, "clientId", {
            type: "Number",
            description: "Strava app client ID",
            default: "0",
        });

        const clientSecret = new CfnParameter(this, "clientSecret", {
            type: "String",
            description: "Strava app client secret",
            default: "",
            noEcho: true,
        });

        const gearIds = new CfnParameter(this, "GearIds", {
            type: "String",
            description: "Stringified JSON of gear IDs to warn about",
            default: '["g9558316"]',
        });

        const role = new Role(this, "LambdaRole", {
            assumedBy: new ServicePrincipal("lambda.amazonaws.com"),
        });
        const ssmPolicy = new PolicyStatement({
            effect: Effect.ALLOW,
            actions: ["ssm:PutParameter", "ssm:GetParameters", "ssm:GetParametersByPath", "ssm:DescribeParameters"],
            resources: [
                ["arn:aws:ssm:", Stack.of(this).region, ":", Stack.of(this).account, ":parameter/strava*"].join(""),
            ],
        });
        role.addToPolicy(ssmPolicy);

        const httpApi = new HttpApi(this, "HttpApi", {
            apiName: "StraveShoesApi",
            createDefaultStage: true,
        });

        const topic = new Topic(this, "Topic", {});
        topic.grantPublish(role);

        const checkLambda = new NodejsFunction(this, "ShoesCheckLambda", {
            role,
            runtime: Runtime.NODEJS_18_X,
            entry: "lib/check-lambda.function.ts",
            memorySize: 128,
            timeout: Duration.seconds(5),
            environment: {
                TOPIC_ARN: topic.topicArn,
                GEAR_IDS: gearIds.valueAsString,
            },
        });

        new Rule(this, "CheckRule", {
            schedule: Schedule.cron({minute: "0", hour: "18", day: "*", month: "*", year: "*"}),
            targets: [
                new LambdaFunction(checkLambda, {
                    retryAttempts: 1,
                    maxEventAge: Duration.minutes(1),
                }),
            ],
        });

        const authLambda = new NodejsFunction(this, "AuthLambda", {
            role,
            runtime: Runtime.NODEJS_18_X,
            entry: "lib/auth-lambda.function.ts",
            memorySize: 128,
            timeout: Duration.seconds(5),
        });
        const authIntegration = new HttpLambdaIntegration("AuthIntegration", authLambda);
        httpApi.addRoutes({
            path: "/auth",
            methods: [HttpMethod.GET],
            integration: authIntegration,
        });

        new StringParameter(this, "ClientIdParameter", {
            parameterName: "/strava/clientId",
            tier: ParameterTier.STANDARD,
            stringValue: clientId.valueAsString,
            dataType: ParameterDataType.TEXT,
        });

        new StringParameter(this, "ClientSecretParameter", {
            parameterName: "/strava/clientSecret",
            tier: ParameterTier.STANDARD,
            stringValue: clientSecret.valueAsString,
            dataType: ParameterDataType.TEXT,
        });

        new StringParameter(this, "AccessTokenParameter", {
            parameterName: "/strava/accessToken",
            tier: ParameterTier.STANDARD,
            stringValue: "placeholder",
            dataType: ParameterDataType.TEXT,
        });

        new StringParameter(this, "RefreshTokenParameter", {
            parameterName: "/strava/refreshToken",
            tier: ParameterTier.STANDARD,
            stringValue: "placeholder",
            dataType: ParameterDataType.TEXT,
        });

        new CfnOutput(this, "AuthCallbackDomain", {
            exportName: "AuthCallbackDomain",
            value: httpApi.apiEndpoint.replace("https://", ""),
        });

        new CfnOutput(this, "StravaAuthUrl", {
            exportName: "StravaAuthUrl",
            value: [
                "https://www.strava.com/oauth/authorize?client_id=",
                clientId.valueAsString,
                "&response_type=code&redirect_uri=",
                httpApi.url,
                "auth&scope=activity:read",
            ].join(""),
        });
    }
}
