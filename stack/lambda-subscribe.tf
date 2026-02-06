module "lambda_subscribe" {
  source = "github.com/ockendenjo/tfmods//lambda"

  aws_env                  = var.env
  name                     = "subscribe"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "strava"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["subscribe"]

  environment = {
    CALLBACK_URL = "${aws_apigatewayv2_stage.default.invoke_url}/event"
  }
}

module "iam_ssm_lambda_subscribe" {
  source      = "github.com/ockendenjo/tfmods//iam-ssm"
  role_id     = module.lambda_subscribe.role_id
  ssm_arn     = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/strava*"
  allow_write = true
}
