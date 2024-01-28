# External secret validator

AWS Lambda function to check K8S `ExternalSecret` definitions VS actual AWS account: ensure that secrets mentioned exists, they have valid json structure and keys referenced are actually present.

## Usage

```bash
helm template . | yq e '. | select(.kind == "ExternalSecret")' > ExternalSecret.yaml
curl --fail-with-body -v --data-binary @ExternalSecret.yaml https://xxx.lambda-url.us-east-1.on.aws\?region\=us-east-1
```

## Deploy

```bash
GOOS=linux GOARCH=amd64 go build -o main .
zip lambda-handler.zip main
aws lambda update-function-code --function ilia-test --zip-file fileb://lambda-handler.zip
```
