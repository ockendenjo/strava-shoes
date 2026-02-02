resource "aws_cloudwatch_event_rule" "gear_check_schedule" {
  name                = "strava-gear-check-schedule"
  description         = "Run gear check daily at 18:00"
  schedule_expression = "cron(0 18 * * ? *)"
}

resource "aws_cloudwatch_event_target" "gear_check_lambda" {
  rule      = aws_cloudwatch_event_rule.gear_check_schedule.name
  target_id = "GearCheckLambda"
  arn       = module.lambda_gear_check.arn

  retry_policy {
    maximum_event_age_in_seconds = 60
    maximum_retry_attempts       = 1
  }
}

resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_gear_check.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.gear_check_schedule.arn
}
