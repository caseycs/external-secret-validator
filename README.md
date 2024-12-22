# External secret validator

AWS Lambda function to check K8S `ExternalSecret` definitions VS actual AWS account: ensure that secrets exists, their json structure is valid and contains referenced keys.

Useful to validate Helm charts or K8S definitions before deployment, to catch typos, missing keys or secrets earlier and explain the issue in human-readable way.

It seems way easier to make make CI perform these checks, but that will mean granting workflow access to read secret values, which is barely acceptable for production deployments.

## Basic usage

```bash
URL=$(aws lambda get-function-url-config --region us-east-1 --function-name external-secret-validator --query 'FunctionUrl' --output text)
echo "External secrets validator url: $URL"
helm template -f values.yaml . | yq '. | select(.kind == "ExternalSecret")' | tee externalSecret.yaml
if [ -s "externalSecret.yaml" ] ; then curl --fail-with-body --data-binary @externalSecret.yaml $URL?region=us-east-1; fi
```

Example output:

```
External secrets validator url: https://xxx.lambda-url.us-east-1.on.aws
Secret found: project-staging-app1-database
Secret key found: project-staging-app1-database/postgres_dsn
Secret found: project-staging-app1-redis
Secret key NOT found: project-staging-app1-redis/redis_dsn: key not found
curl: (22) The requested URL returned error: 400
```

## Installation using Terragrunt

```hcl
include "root" {
  path = find_in_parent_folders()
}

include "provider" {
  path = find_in_parent_folders("provider-aws.hcl")
}

terraform {
  source = "git@github.com:caseycs/external-secret-validator.git//terraform?ref=v0.0.7"
}

# default values, could be adjusted or ommited
inputs = {
  function_name = "external-secret-validator"
  policy_json = <<END
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
```

### IAM policy for CI role/user

In order CI will be able to get function url by name you need to add the following policy to it's IAM role/user:

```json
{
  "Statement": [
      {
          "Action": [
              "lambda:GetFunctionUrlConfig"
          ],
          "Effect": "Allow",
          "Resource": "arn:aws:lambda:us-east-1:xxx:function:external-secret-validator",
          "Sid": "ExternalSecretValidator"
      }
  ]
}
```

## Build and update existing Lambda manually

```bash
GOOS=linux GOARCH=amd64 go build -o main .
zip lambda-handler.zip main
aws lambda update-function-code --function ilia-test --zip-file fileb://lambda-handler.zip
```
