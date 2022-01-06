# strava-shoes

This project contains source code for an AWS serverless application for checking that activities on Strava have the correct gear (footwear, bike etc.) set. 

## Longer intro

This project contains a Serverless Application Model (SAM) stack which configures a number of lambda functions. One lambda
function is used to handle the authorization response from strava. The other lambda function is run on a schedule
(via CloudWatch events) and queries the Strava API to check the gear assigned to activities.

Subscriptions can be added to the configured SNS topic to receive notifications.

## Strava App

Before creating the AWS stack, create a Strava app. This is required for the Lambda function to be able to access activity data from your strava account.

Use `example.com` as the **Authorization Callback Domain** for now. 

## Deploy

To build and deploy your application for the first time, run the following in your shell:

```bash
sam build
sam deploy --guided
```

The first command will build the source of your application. The second command will package and deploy your application to AWS, with a series of prompts:

* **Stack Name**: The name of the stack to deploy to CloudFormation. This should be unique to your account and region, and a good starting point would be something matching your project name.
* **AWS Region**: The AWS region you want to deploy your app to.
* **ClientId**: Strava App client ID (from the app details on Strava)
* **ClientSecret**: Strava App client secret (from the app details on Strava)
* **AthleteId**: Your strava athlete ID
* **GearIds**: Any gear which should trigger a notification - e.g. `["g1234", null]`

Note: The ClientSecret would be better stored as a `SecureString` parameter in systems manager, but CloudFormation doesn't
support that functionality. For better security the client secret could be removed as a stack parameter
and entered into the parameter store manually. Or, use SecretsManager instead, but this was intended to be a low-cost solution.

## Stack outputs

The stack produces two outputs: 
* **AuthCallbackDomain** - this needs to be copied into the Strava app settings
* **StravaAuthUrl** - visit this URL to authorize access to your Strava activities

## Cleanup

To delete the sample application that you created, use the AWS CLI. Assuming you used your project name for the stack name, you can run the following:

```bash
sam delete
```
or
```bash
aws cloudformation delete-stack --stack-name strava-shoes
```
