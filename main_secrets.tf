resource "aws_secretsmanager_secret" "client_passphrase" {
  count = var.use_aws_secrets_manager ? 1 : 0

  name = "${var.aws_lambda_function_function_name}-client_passphrase"
}

resource "aws_secretsmanager_secret_version" "client_passphrase" {
  count = var.use_aws_secrets_manager ? 1 : 0

  secret_id     = aws_secretsmanager_secret.client_passphrase[0].id
  secret_string = var.client_passphrase
}

resource "aws_secretsmanager_secret" "issuer_passphrase" {
  count = var.use_aws_secrets_manager ? 1 : 0

  name = "${var.aws_lambda_function_function_name}-issuer_passphrase"
}

resource "aws_secretsmanager_secret_version" "issuer_passphrase" {
  count = var.use_aws_secrets_manager ? 1 : 0

  secret_id     = aws_secretsmanager_secret.issuer_passphrase[0].id
  secret_string = var.issuer_passphrase
}
