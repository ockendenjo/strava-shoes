module "lambda_receive_event" {
  source = "github.com/ockendenjo/tfmods//lambda"

  aws_env                  = var.env
  name                     = "receive-event"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "strava"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["receive-event"]

  environment = {}
}

module "iam_eventbridge_lambda_receive_event" {
  source  = "github.com/ockendenjo/tfmods//iam-eventbridge"
  role_id = module.lambda_receive_event.role_id
  bus_arns = [
    "arn:aws:events:${var.aws_region}:${var.aws_account_id}:event-bus/default"
  ]
}
