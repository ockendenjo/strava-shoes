module "lambda_gear_check" {
  source = "github.com/ockendenjo/tfmods//lambda"

  aws_env                  = var.env
  name                     = "gear-check"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "strava"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["check"]

  environment = {
    GEAR_IDS   = var.gear_ids
    TOPIC_ARN  = aws_sns_topic.topic.arn
    BAGGING_DB = aws_dynamodb_table.bagging_db.name
  }
}

module "iam_ssm_lambda_check" {
  source      = "github.com/ockendenjo/tfmods//iam-ssm"
  role_id     = module.lambda_gear_check.role_id
  ssm_arn     = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/strava*"
  allow_write = true
}

module "iam_dynamodb_lambda_check" {
  source = "github.com/ockendenjo/tfmods//iam-dynamodb"
  dynamo_table_arns = [
    aws_dynamodb_table.bagging_db.arn,
  ]
  role_id = module.lambda_gear_check.role_id
}

module "iam_sns_lambda_check" {
  source  = "github.com/ockendenjo/tfmods//iam-sns"
  role_id = module.lambda_gear_check.role_id
  sns_arns = [
    aws_sns_topic.topic.arn,
  ]
}

module "iam_eventbridge_lambda_check" {
  source  = "github.com/ockendenjo/tfmods//iam-eventbridge"
  role_id = module.lambda_gear_check.role_id
  bus_arns = [
    "arn:aws:events:${var.aws_region}:${var.aws_account_id}:event-bus/default"
  ]
}
