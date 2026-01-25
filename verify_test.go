package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// mockSecretsManagerClient is a mock implementation of SecretsManagerAPI for testing.
type mockSecretsManagerClient struct {
	secrets map[string]string // map of secret name to secret JSON value
}

func (m *mockSecretsManagerClient) ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	var secretList []types.SecretListEntry

	// Find secrets matching the filter
	for _, filter := range params.Filters {
		if filter.Key == "name" {
			for _, filterValue := range filter.Values {
				if _, exists := m.secrets[filterValue]; exists {
					secretList = append(secretList, types.SecretListEntry{
						Name: aws.String(filterValue),
						ARN:  aws.String(fmt.Sprintf("arn:aws:secretsmanager:us-east-1:123456789012:secret:%s", filterValue)),
					})
				}
			}
		}
	}

	return &secretsmanager.ListSecretsOutput{
		SecretList: secretList,
	}, nil
}

func (m *mockSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	secretID := aws.ToString(params.SecretId)

	// Extract secret name from ARN if needed
	secretName := secretID
	if strings.HasPrefix(secretID, "arn:aws:secretsmanager:") {
		parts := strings.Split(secretID, ":")
		if len(parts) >= 7 {
			secretName = parts[6]
		}
	}

	if value, exists := m.secrets[secretName]; exists {
		return &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(value),
		}, nil
	}

	return nil, fmt.Errorf("secret not found: %s", secretName)
}

// newMockClient creates a mock client with test data.
func newMockClient() *mockSecretsManagerClient {
	return &mockSecretsManagerClient{
		secrets: map[string]string{
			"external_secret_validator_test": `{"int_key": 123, "string_key": "test_value"}`,
		},
	}
}

func TestVerifyExternalSecretSuccess(t *testing.T) {
	mockClient := newMockClient()

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

			output, errorsFound, err := verifyExternalSecretYamlWithClient(yamlFile, mockClient)

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
	mockClient := newMockClient()

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

			_, _, err = verifyExternalSecretYamlWithClient(yamlFile, mockClient)

			if err == nil {
				t.Errorf("Expected error, but got none")
			}

			if err.Error() != testCase.errorText {
				t.Errorf("Expected error text: %s, but got %s", testCase.errorText, err.Error())
			}
		})
	}
}
