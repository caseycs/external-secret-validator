variable "function_name" {
  default = "external-secret-validator"
}

variable "function_version" {
  default = "v0.0.3"
}

variable "policy_json" {
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
