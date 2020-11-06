################################################################################
# main_iam.tf
################################################################################

# aws_iam_policy.letsencrypt-lambda_policy
variable "aws_iam_policy_name" {
  default = "letsencrypt-lambda_policy"
}

variable "aws_iam_policy_path" {
  default = "/"
}

variable "aws_iam_policy_description" {
  default = "letsencrypt policy"
}

variable "aws_iam_policy_additional_statements" {
  type = list

  default = []
}

# aws_iam_role.letsencrypt-lambda_role
variable "aws_iam_role_name" {
  default = "letsencrypt-lambda_role"
}
