# external secret validator

## deploy

```bash
GOOS=linux GOARCH=amd64 go build -o main .
zip lambda-handler.zip main
aws lambda update-function-code --function ilia-test --zip-file fileb://lambda-handler.zip
```