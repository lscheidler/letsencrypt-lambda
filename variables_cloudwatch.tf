################################################################################
# main_cloudwatch.tf
################################################################################

variable "use_cloudwatch_event" {
  type = bool

  default = true
}

# aws_cloudwatch_event_target.letsencrypt-lambda
variable "aws_cloudwatch_event_target_target_id" {
  default = ""
}

# aws_cloudwatch_event_rule.letsencrypt-lambda
variable "aws_cloudwatch_event_rule_name" {
  default = ""
}

variable "aws_cloudwatch_event_rule_description" {
  default = ""
}

variable "schedule_expression" {
  default = "cron(01 03 * * ? *)"
}
