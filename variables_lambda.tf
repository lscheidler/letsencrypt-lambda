################################################################################
# main_lambda.tf
################################################################################

# aws_lambda_function.letsencrypt-lambda
variable "aws_lambda_function_function_name" {
  default = "letsencrypt-lambda"
}

variable "aws_lambda_function_publish" {
  default = true
}

variable "aws_region" {
  default = ""
}

variable "aws_assume_role" {
  default = ""
}

variable "aws_hosted_zone_id" {}
variable "client_passphrase" {}
variable "domains" {}

variable "dynamodb_table_name" {
  default = "LetsencryptCA"
}

variable "email" {}
variable "issuer_passphrase" {}

# aws_lambda_alias.letsencrypt-lambda
variable "aws_lambda_alias_name" {
  default = "dev"
}

variable "aws_lambda_alias_description" {
  default = "letsencrypt-lambda dev"
}

# aws_lambda_function_event_invoke_config.letsencrypt-lambda
variable "on_failure" {
  default = []
}

variable "on_success" {
  default = []
}
