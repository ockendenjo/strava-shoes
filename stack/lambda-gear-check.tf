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
