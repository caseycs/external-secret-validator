package main

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	testCases := []struct {
		name                string
		filename            string
		expectedError       bool
		errorText           string
		expectedErrorsFound int
		outputSubstrings    []string
	}{
		{
			name:                "Empty file",
			filename:            "test/externalsecret/empty.yaml",
			expectedError:       false,
			expectedErrorsFound: 0,
			outputSubstrings:    []string{"Empty YAML"},
		},
		{
			name:                "Invalid yaml",
			filename:            "test/externalsecret/invalid-yaml.yaml",
			expectedError:       true,
			errorText:           "Unexpected kind",
			expectedErrorsFound: 0,
		},
		{
			name:                "No data",
			filename:            "test/externalsecret/no-data.yaml",
			expectedError:       false,
			expectedErrorsFound: 0,
			outputSubstrings:    []string{"Empty .Spec.Data"},
		},
		{
			name:                "Other kind",
			filename:            "test/externalsecret/other-kind.yaml",
			expectedError:       true,
			errorText:           "Unexpected kind",
			expectedErrorsFound: 0,
		},
		{
			name:                "Secrets found, keys found",
			filename:            "test/externalsecret/success.yaml",
			expectedError:       false,
			expectedErrorsFound: 0,
		},
		{
			name:                "Secret not found",
			filename:            "test/externalsecret/non-existing-secret.yaml",
			expectedError:       false,
			expectedErrorsFound: 1,
			outputSubstrings:    []string{"Secret NOT found"},
		},
		{
			name:                "Secret key not found",
			filename:            "test/externalsecret/non-existing-key.yaml",
			expectedError:       false,
			expectedErrorsFound: 1,
			outputSubstrings:    []string{"Secret key NOT found"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			yamlFile, err := os.ReadFile(testCase.filename)

			if err != nil {
				log.Fatalf("Error reading YAML file: %v", err)
			}

			output, errorsFound, err := verifyExternalSecretYaml(yamlFile, "us-east-1")

			if testCase.expectedError == false && err != nil {
				t.Errorf("Expected no error, but got %v, output: %s", err, output)
			} else if testCase.expectedError == true {
				if err == nil {
					t.Errorf("Expected error, but got none, output: %s", output)
				} else if err.Error() != testCase.errorText {
					t.Errorf("Expected error text: %s, but got %s", testCase.errorText, err.Error())
				}
			}

			if errorsFound != testCase.expectedErrorsFound {
				t.Errorf("Expected errors found %d, but got %d, output: %s", testCase.expectedErrorsFound, errorsFound, output)
			}

			for _, v := range testCase.outputSubstrings {
				if !strings.Contains(string(output), v) {
					t.Errorf("Output substring missing: %s, output: %s", v, output)
				}
			}
		})
	}
}
