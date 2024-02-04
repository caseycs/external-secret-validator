package main

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestVerifyExternalSecretSuccess(t *testing.T) {
	testCases := []struct {
		name             string
		filename         string
		errorsFound      int
		outputSubstrings []string
	}{{
		name:             "No data",
		filename:         "test/externalsecret/no-data.yaml",
		errorsFound:      0,
		outputSubstrings: []string{"Empty .Spec.Data"},
	},
		{
			name:        "Secrets found, keys found",
			filename:    "test/externalsecret/success.yaml",
			errorsFound: 0,
		},
		{
			name:             "Secret not found",
			filename:         "test/externalsecret/non-existing-secret.yaml",
			errorsFound:      1,
			outputSubstrings: []string{"Secret NOT found"},
		},
		{
			name:             "Secret key not found",
			filename:         "test/externalsecret/non-existing-key.yaml",
			errorsFound:      1,
			outputSubstrings: []string{"Secret key NOT found"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			yamlFile, err := os.ReadFile(testCase.filename)

			if err != nil {
				log.Fatalf("Error reading YAML file: %v", err)
			}

			output, errorsFound, err := verifyExternalSecretYaml(yamlFile, "us-east-1")

			if err != nil {
				t.Errorf("Expected no error, but got %v, output: %s", err, output)
			}

			if errorsFound != testCase.errorsFound {
				t.Errorf("Expected errors found %d, but got %d, output: %s", testCase.errorsFound, errorsFound, output)
			}

			for _, v := range testCase.outputSubstrings {
				if !strings.Contains(string(output), v) {
					t.Errorf("Output substring missing: %s, output: %s", v, output)
				}
			}
		})
	}
}

func TestVerifyExternalSecretError(t *testing.T) {
	testCases := []struct {
		name      string
		filename  string
		errorText string
	}{
		{
			name:      "Empty file",
			filename:  "test/externalsecret/empty.yaml",
			errorText: "Empty YAML",
		},
		{
			name:      "Invalid yaml",
			filename:  "test/externalsecret/invalid-yaml.yaml",
			errorText: "Unexpected kind",
		},
		{
			name:      "Other kind",
			filename:  "test/externalsecret/other-kind.yaml",
			errorText: "Unexpected kind",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			yamlFile, err := os.ReadFile(testCase.filename)

			if err != nil {
				log.Fatalf("Error reading YAML file: %v", err)
			}

			_, _, err = verifyExternalSecretYaml(yamlFile, "us-east-1")

			if err == nil {
				t.Errorf("Expected error, but got none")
			}

			if err.Error() != testCase.errorText {
				t.Errorf("Expected error text: %s, but got %s", testCase.errorText, err.Error())
			}
		})
	}
}
