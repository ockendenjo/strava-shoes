# strava-shoes

This project contains source code for an AWS serverless application for checking that activities on Strava have the correct gear (footwear, bike etc.) set. 

## Longer intro

This project contains a Terraform stack which configures a number of lambda functions. 
One lambda function is used to handle the authorization response from strava. 
The other lambda function is run on a schedule (via CloudWatch events) and queries the Strava API to check the gear 
assigned to activities.

Subscriptions can be added to the configured SNS topic to receive notifications.

## Strava App

Before creating the AWS stack, create a Strava app. This is required for the Lambda function to be able to access activity data from your strava account.

Use `example.com` as the **Authorization Callback Domain** for now. 

## Deploy

* Modify the `upload-cmd` task to upload lambda binaries to your own bucket.
* Modify the `tfvars/pro.tfvars` file to use your own resources.
* Modify the `tfvars/backend-pro.hcl` file to use your own bucket.
* Run `xc init` to initialise the project.
* Run `xc apply` to deploy the stack.

## Stack outputs

The stack produces two outputs: 
* **AuthCallbackDomain** - this needs to be copied into the Strava app settings
* **StravaAuthUrl** - visit this URL to authorize access to your Strava activities

## Cleanup

Use `terraform destroy -auto-approve`

## tasks

[xcfile.dev](xcfile.dev) tasks:

### apply

requires: upload-cmd, just-apply

### build-cmd

requires: clean

```shell
go run ./scripts/build-cmd --zip
```

### clean

```shell
rm -rf build/* || true
```

### format

directory: stack

```shell
terraform fmt --recursive --write .
```

### init

directory: stack
environment: AWS_PROFILE=strava

```shell
terraform init -backend-config="tfvars/backend-pro.hcl" -reconfigure
```

### just-apply

directory: stack
environment: AWS_PROFILE=strava

```shell
terraform apply -auto-approve -var-file="tfvars/pro.tfvars"
```

### plan

requires: upload-cmd
directory: stack
environment: AWS_PROFILE=strava

```shell
terraform plan -var-file="tfvars/pro.tfvars"
```

### profile

```shell
cat >> ~/.aws/config <<EOF
[profile strava]
region = eu-west-1
output = json
role_arn = arn:aws:iam::574363388371:role/strava-cicd
source_profile = default

EOF
```

### upload-cmd

requires: build-cmd
environment: AWS_PROFILE=strava
environment: BINARY_BUCKET=strava-lambda-binaries20260202095142751100000001

```shell
go run ./scripts/upload-binaries
```
