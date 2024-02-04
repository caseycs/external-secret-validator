variable "function_name" {
  type    = string
  default = "external-secret-validator"
}

variable "package_url" {
  type    = string
  default = "https://github.com/caseycs/external-secret-validator/releases/download/v0.0.5/lambda-handler.zip"
}

variable "policy_json" {
  type        = string
  description = "IAM policy to access secrets"
  default     = <<END
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetRandomPassword",
                "secretsmanager:GetResourcePolicy",
                "secretsmanager:GetSecretValue",
                "secretsmanager:DescribeSecret",
                "secretsmanager:ListSecretVersionIds",
                "secretsmanager:ListSecrets",
                "secretsmanager:BatchGetSecretValue"
            ],
            "Resource": "*"
        }
    ]
}
  END
}
