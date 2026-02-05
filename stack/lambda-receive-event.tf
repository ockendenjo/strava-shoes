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
