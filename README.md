# letsencrypt-lambda

AWS lambda function for creating and updating letsencrypt certificates

## Usage

You can use terraform (>= 0.13.5) to deploy the lambda function:

```
module "letsencrypt-lambda" {
  source = "github.com/lscheidler/letsencrypt-lambda?ref=main"

  email              = "me@example.com"
  domains            = "example.com,*.example.com"
  aws_hosted_zone_id = "Z123ABC456DEF7"
  issuer_passphrase  = "<secure_issuer_passphrase>"
  client_passphrase  = "<secure_client_passphrase>"

  on_failure = [data.aws_sns_topic.topic.arn]

  aws_iam_policy_additional_statements = [
    {
      effect = "Allow",
      actions = [
        "sns:Publish",
      ],
      resources = [
        data.aws_sns_topic.topic.arn,
      ]
    }
  ]
}
```

It is going to configure
- iam role and policy for required permissions
- lambda function
- secrets (issuer\_passphrase, client\_passphrase) to secrets manager (optional)
- cloudwatch event rule to run lambda daily

## Argument Reference

| Name                                    | Required  | Default                                     | Description                                     |
|-----------------------------------------|-----------|---------------------------------------------|-------------------------------------------------|
| `aws_hosted_zone_id`                    | 🗹         |                                             | Route53 Domain id                               |
| `client_passphrase`                     | 🗹         |                                             | Client passphrase for certificate encryption    |
| `domains`                               | 🗹         |                                             | Domains to get a certificate for                |
| `email`                                 | 🗹         |                                             | Registration email for letsencrypt              |
| `issuer_passphrase`                     | 🗹         |                                             | Issuer passphrase for letsencrypt account data  |
| `aws_region`                            | 🗷         | `""`                                        |                                                 |
| `aws_assume_role`                       | 🗷         | `""`                                        |                                                 |
| `aws_iam_policy_name`                   | 🗷         | `"letsencrypt-lambda_policy"`               |                                                 |
| `aws_iam_policy_path`                   | 🗷         | `"/"`                                       |                                                 |
| `aws_iam_policy_description`            | 🗷         | `"letsencrypt policy"`                      |                                                 |
| `aws_iam_policy_additional_statements`  | 🗷         | `[]`                                        |                                                 |
| `aws_iam_role_name`                     | 🗷         | `"letsencrypt-lambda_role"`                 |                                                 |
| `aws_lambda_function_function_name`     | 🗷         | `"letsencrypt-lambda"`                      |                                                 |
| `aws_lambda_function_publish`           | 🗷         | `true`                                      |                                                 |
| `aws_lambda_alias_name`                 | 🗷         | `"dev"`                                     |                                                 |
| `aws_lambda_alias_description`          | 🗷         | `"letsencrypt-lambda dev"`                  |                                                 |
| `dynamodb_table_name`                   | 🗷         | `"LetsencryptCA"`                           |                                                 |
| `use_aws_secrets_manager`               | 🗷         | `true`                                      |                                                 |
| `use_cloudwatch_event`                  | 🗷         | `true`                                      |                                                 |
| `aws_cloudwatch_event_target_target_id` | 🗷         | `""` => `aws_lambda_function_function_name` |                                                 |
| `aws_cloudwatch_event_rule_name`        | 🗷         | `""` => `aws_lambda_function_function_name` |                                                 |
| `aws_cloudwatch_event_rule_description` | 🗷         | `""` => `aws_lambda_function_function_name` |                                                 |
| `schedule_expression`                   | 🗷         | `"cron(01 03 * * ? *)"`                     |                                                 |

## License

The lambda function is available as open source under the terms of the [Apache 2.0 License](http://opensource.org/licenses/Apache-2.0).
