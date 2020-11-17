resource "aws_iam_policy" "letsencrypt-lambda_policy" {
  name        = var.aws_iam_policy_name
  path        = var.aws_iam_policy_path
  description = var.aws_iam_policy_description

  policy = data.aws_iam_policy_document.letsencrypt-lambda_policy.json
}

data "aws_iam_policy_document" "letsencrypt-lambda_policy" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      #"arn:aws:logs:*:*:*"
      "arn:aws:logs:*:*:log-group:/aws/lambda/${var.aws_lambda_function_function_name}:*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "dynamodb:CreateTable",
      "dynamodb:GetItem",
      "dynamodb:UpdateItem",
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${var.dynamodb_table_name}",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "route53:ChangeResourceRecordSets",
    ]
    resources = [
      "arn:aws:route53:::hostedzone/${var.aws_hosted_zone_id}",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "route53:GetChange",
    ]
    resources = [
      "*",
    ]
  }

  dynamic "statement" {
    for_each = var.use_aws_secrets_manager ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "secretsmanager:GetSecretValue",
      ]
      resources = [
        aws_secretsmanager_secret.client_passphrase[0].arn,
        aws_secretsmanager_secret.issuer_passphrase[0].arn,
      ]
    }
  }

  dynamic "statement" {
    for_each = var.aws_iam_policy_additional_statements

    content {
      effect    = statement.value["effect"]
      actions   = statement.value["actions"]
      resources = statement.value["resources"]
    }
  }
}

resource "aws_iam_role" "letsencrypt-lambda_role" {
  name = var.aws_iam_role_name

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_iam_role_policy_attachment" "attach" {
  role       = aws_iam_role.letsencrypt-lambda_role.name
  policy_arn = aws_iam_policy.letsencrypt-lambda_policy.arn
}

