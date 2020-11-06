resource "aws_cloudwatch_event_target" "letsencrypt-lambda" {
  count = var.use_cloudwatch_event ? 1 : 0

  target_id = local.aws_cloudwatch_event_target_target_id
  rule      = aws_cloudwatch_event_rule.letsencrypt-lambda[0].name
  arn       = aws_lambda_function.letsencrypt-lambda.arn
}

resource "aws_cloudwatch_event_rule" "letsencrypt-lambda" {
  count = var.use_cloudwatch_event ? 1 : 0

  name        = local.aws_cloudwatch_event_rule_name
  description = local.aws_cloudwatch_event_rule_description

  # UTC
  schedule_expression = var.schedule_expression
}

resource "aws_lambda_permission" "letsencrypt-lambda_allow_cloudwatch" {
  count = var.use_cloudwatch_event ? 1 : 0

  statement_id  = "AlarmDowntimeAllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.letsencrypt-lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.letsencrypt-lambda[0].arn
}
