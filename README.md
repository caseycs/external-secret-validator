# External secret validator

AWS Lambda function to check K8S `ExternalSecret` definitions VS actual AWS account: ensure that secrets exists, they json structure is valid and contains referenced keys.

Useful to validate Helm charts or K8S definitions before deployment, to catch typos, missing keys or secrets earlier and explain the issue in human-readable way.

It seems way easier to make make CI perform these checks, but that will mean granting workflow access to read secret values, which is barely acceptable for production deployments.

## Basic usage

```bash
helm template -f values.yaml . | yq '. | select(.kind == "ExternalSecret")' | tee externalSecret.yaml
if [ -s "externalSecret.yaml" ] ; then curl --fail-with-body --data-binary @externalSecret.yaml https://xxx.lambda-url.us-east-1.on.aws; fi
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
  source = "git@github.com:caseycs/external-secret-validator.git//terraform?ref=v0.0.5"
}

# Could be ommited, but you probably do not want to use external package
# that will access to your aws secrets on prod, so makeing a private fork
# of the repo and redefining package_url sounds like a good idea
inputs = {
  package_url = "https://github.com/caseycs/external-secret-validator/releases/download/v0.0.5/lambda-handler.zip"
}
```

## Build and update Lambda

```bash
GOOS=linux GOARCH=amd64 go build -o main .
zip lambda-handler.zip main
aws lambda update-function-code --function ilia-test --zip-file fileb://lambda-handler.zip
```
