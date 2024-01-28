package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaStart(ctx context.Context, event *events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// http.HandleFunc("/", getRoot)
	// log.Printf("Listening on 3333")
	// err := http.ListenAndServe(":3333", nil)
	// if err != nil {
	// 	log.Fatalf("Failed to start http server, %v", err)
	// }

	// yml, err := os.ReadFile("config.yaml")
	// if err != nil {
	// 	log.Fatalf("Error reading YAML file: %v", err)
	// }

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

	log.Print(string(yml))
	region := "us-east-1"

	log, errors, err := verifyExternalSecretYaml(yml, region)

	// handle function result
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
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(lambdaStart)
}
