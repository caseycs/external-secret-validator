# External secret validator

AWS Lambda function to check K8S `ExternalSecret` definitions VS actual AWS account: ensure that secrets exists, their json structure is valid and contains referenced keys.

Useful to validate Helm charts or K8S definitions before deployment, to catch typos, missing keys or secrets earlier and explain the issue in human-readable way.

It seems way easier to make make CI perform these checks, but that will mean granting workflow access to read secret values, which is barely acceptable for production deployments.

## Basic usage

```bash
helm template -f values.yaml . | yq '. | select(.kind == "ExternalSecret")' | tee externalSecret.yaml
if [ -s "externalSecret.yaml" ] ; then curl --fail-with-body --data-binary @externalSecret.yaml https://xxx.lambda-url.us-east-1.on.aws; fi
```

Example output:

```
Secret found: project-staging-app-redis
Secret key NOT found: project-staging-app-redis/redis_addresses: key not found
curl: (22) The requested URL returned error: 400
```

## Install via Terragrunt

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

## Build and update existing Lambda manually

```bash
GOOS=linux GOARCH=amd64 go build -o main .
zip lambda-handler.zip main
aws lambda update-function-code --function ilia-test --zip-file fileb://lambda-handler.zip
```
