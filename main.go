package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaStart(ctx context.Context, event *events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// check for empty body
	if event.Body == "" {
		return events.LambdaFunctionURLResponse{
			Body:       "Non-empty request body expected",
			StatusCode: 400,
		}, nil
	}

	// base64 decode if necessary
	yml := []byte(event.Body)
	if event.IsBase64Encoded {
		ymlFromBase64, err := base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			return events.LambdaFunctionURLResponse{
				Body:       fmt.Sprintf("Error decoding body: %v ", err.Error()),
				StatusCode: 500,
			}, nil
		}
		yml = ymlFromBase64
	}

	// parse region
	region, regionFound := event.QueryStringParameters["region"]
	if !regionFound {
		region = "us-east-1"
	}

	log, errors, err := verifyExternalSecretYaml(yml, region)

	// handle verification result
	if err != nil {
		return events.LambdaFunctionURLResponse{
			Body:       err.Error(),
			StatusCode: 400,
		}, nil
	}

	statusCode := 200
	if errors > 0 {
		statusCode = 400
	}

	return events.LambdaFunctionURLResponse{
		Body:       string(log),
		StatusCode: statusCode,
	}, nil
}

func main() {
	lambda.Start(lambdaStart)
}
