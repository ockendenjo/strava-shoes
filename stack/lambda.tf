resource "aws_lambda_function" "gear_check" {
  filename         = "build/check/bootstrap.zip"
  function_name    = "Strava_GearCheckLambda"
  role             = aws_iam_role.gear_check_lambda_role.arn
  handler          = "function"
  source_code_hash = filebase64sha256("build/check/bootstrap.zip")
  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  timeout          = 10
  memory_size      = 512

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      GEAR_IDS   = var.gear_ids
      TOPIC_ARN  = aws_sns_topic.topic.arn
      BAGGING_DB = aws_dynamodb_table.bagging_db.name
    }
  }
}

resource "aws_lambda_function" "auth" {
  filename         = "build/auth/bootstrap.zip"
  function_name    = "Strava_AuthLambda"
  role             = aws_iam_role.auth_lambda_role.arn
  handler          = "function"
  source_code_hash = filebase64sha256("build/auth/bootstrap.zip")
  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  timeout          = 10
  memory_size      = 512

  tracing_config {
    mode = "Active"
  }
}
