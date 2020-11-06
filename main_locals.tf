locals {
  aws_cloudwatch_event_target_target_id = (var.aws_cloudwatch_event_target_target_id == "") ? var.aws_lambda_function_function_name : var.aws_cloudwatch_event_target_target_id
  aws_cloudwatch_event_rule_name        = (var.aws_cloudwatch_event_rule_name == "") ? var.aws_lambda_function_function_name : var.aws_cloudwatch_event_rule_name
  aws_cloudwatch_event_rule_description = (var.aws_cloudwatch_event_rule_description == "") ? var.aws_lambda_function_function_name : var.aws_cloudwatch_event_rule_description
}
