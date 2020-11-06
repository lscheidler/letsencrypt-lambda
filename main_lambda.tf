data "external" "letsencrypt-lambda" {
  program     = ["bash", "terraform-build.sh"]
  working_dir = path.module
}

resource "aws_lambda_function" "letsencrypt-lambda" {
  filename         = "${path.module}/letsencrypt-lambda.zip"
  function_name    = var.aws_lambda_function_function_name
  role             = aws_iam_role.letsencrypt-lambda_role.arn
  handler          = "letsencrypt-lambda"
  source_code_hash = filebase64sha256(data.external.letsencrypt-lambda.result["filename"])
  runtime          = "go1.x"
  publish          = var.aws_lambda_function_publish

  environment {
    variables = {
      REGION                       = var.aws_region
      ASSUME_ROLE                  = var.aws_assume_role
      AWS_HOSTED_ZONE_ID           = var.aws_hosted_zone_id
      CLIENT_PASSPHRASE            = var.use_aws_secrets_manager ? "" : var.client_passphrase
      CLIENT_PASSPHRASE_SECRET_ARN = var.use_aws_secrets_manager ? aws_secretsmanager_secret.client_passphrase[0].arn : ""
      DOMAINS                      = var.domains
      DYNAMODB_TABLE_NAME          = var.dynamodb_table_name
      EMAIL                        = var.email
      ISSUER_PASSPHRASE            = var.use_aws_secrets_manager ? "" : var.issuer_passphrase
      ISSUER_PASSPHRASE_SECRET_ARN = var.use_aws_secrets_manager ? aws_secretsmanager_secret.issuer_passphrase[0].arn : ""
    }
  }
}

resource "aws_lambda_alias" "letsencrypt-lambda" {
  count = var.aws_lambda_function_publish ? 1 : 0

  name             = "dev"
  description      = "letsencrypt-lambda dev"
  function_name    = aws_lambda_function.letsencrypt-lambda.arn
  function_version = aws_lambda_function.letsencrypt-lambda.version
}

resource "aws_lambda_function_event_invoke_config" "letsencrypt-lambda" {
  count         = var.on_failure != "" || var.on_success != "" ? 1 : 0
  function_name = aws_lambda_alias.letsencrypt-lambda[0].function_name

  destination_config {
    dynamic "on_failure" {
      for_each = var.on_failure

      content {
        destination = on_failure.value
      }
    }

    dynamic "on_success" {
      for_each = var.on_success

      content {
        destination = on_success.value
      }
    }
  }
}
