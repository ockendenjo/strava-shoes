module "lambda_auth" {
  source = "github.com/ockendenjo/tfmods//lambda"

  aws_env                  = var.env
  name                     = "auth"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "strava"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["auth"]

  environment = {}
}

module "iam_ssm_lambda_auth" {
  source  = "github.com/ockendenjo/tfmods//iam-ssm"
  role_id = module.lambda_auth.role_id
  ssm_arn = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/strava*"
}
