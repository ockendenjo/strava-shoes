module "lambda_confirm_sub" {
  source = "github.com/ockendenjo/tfmods//lambda"

  aws_env                  = var.env
  name                     = "confirm-sub"
  permissions_boundary_arn = var.permissions_boundary_arn
  project_name             = "strava"
  s3_bucket                = var.lambda_binaries_bucket
  s3_object_key            = local.manifest["confirm-sub"]

  environment = {}
}

resource "aws_apigatewayv2_integration" "confirm_sub" {
  api_id                 = aws_apigatewayv2_api.http_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = module.lambda_confirm_sub.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "confirm_sub" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "GET /event"
  target    = "integrations/${aws_apigatewayv2_integration.confirm_sub.id}"
}

resource "aws_lambda_permission" "confirm_sub" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_confirm_sub.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}
