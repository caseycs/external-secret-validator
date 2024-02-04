package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type ExternalSecret struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		RefreshInterval string `yaml:"refreshInterval"`
		SecretStoreRef  struct {
			Name string `yaml:"name"`
			Kind string `yaml:"kind"`
		} `yaml:"secretStoreRef"`
		Target struct {
			Name           string `yaml:"name"`
			CreationPolicy string `yaml:"creationPolicy"`
		} `yaml:"target"`
		Data []struct {
			SecretKey string `yaml:"secretKey"`
			RemoteRef struct {
				Key      string `yaml:"key"`
				Property string `yaml:"property"`
			} `yaml:"remoteRef"`
		} `yaml:"data"`
	} `yaml:"spec"`
}

func verifyExternalSecretYaml(yml []byte, region string) ([]byte, int, error) {
	// basic checks
	if strings.TrimSpace(string(yml)) == "" {
		return  nil, 0, fmt.Errorf("Empty YAML")
	}

	var externalSecret ExternalSecret
	err := yaml.Unmarshal(yml, &externalSecret)
	if err != nil {
		return nil, 0, fmt.Errorf("Error unmarshalling YAML: %v", err)
	}

	if externalSecret.Kind != "ExternalSecret" {
		return nil, 0, fmt.Errorf("Unexpected kind")
	}

	if len(externalSecret.Spec.Data) == 0 {
		return []byte("Empty .Spec.Data"), 0, nil
	}

	// check via aws api
	ctx := context.Background()

	// aws sm client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to load AWS configuration, %v", err)
	}
	client := secretsmanager.NewFromConfig(cfg)

	b := &bytes.Buffer{}
	errors := 0

	// loop over references found
	for _, data := range externalSecret.Spec.Data {
		// secret
		secret, err := findSecretByName(ctx, client, data.RemoteRef.Key)
		if err != nil {
			errors++
			fmt.Fprintf(b, "Secret NOT found: %s\n", data.RemoteRef.Key)
			continue
		}
		fmt.Fprintf(b, "Secret found: %s\n", aws.ToString(secret.Name))

		secretValue, err := getSecretValue(ctx, client, aws.ToString(secret.ARN))
		if err != nil {
			fmt.Fprintf(b, "Error getting secret value: %v\n", err)
			continue
		}

		// secret key
		err2 := checkJsonKeyString([]byte(secretValue), data.RemoteRef.Property)
		if err2 != nil {
			errors++
			fmt.Fprintf(b, "Secret key NOT found: %s/%s: %v\n", aws.ToString(secret.Name), data.RemoteRef.Property, err2)
		} else {
			fmt.Fprintf(b, "Secret key found: %s/%s\n", aws.ToString(secret.Name), data.RemoteRef.Property)
		}
	}

	return b.Bytes(), errors, nil
}

func findSecretByName(ctx context.Context, client *secretsmanager.Client, secretName string) (*types.SecretListEntry, error) {
	// Create an input for ListSecrets.
	input := &secretsmanager.ListSecretsInput{
		Filters: []types.Filter{
			{
				Key:    "name",
				Values: []string{secretName},
			},
		},
	}

	// Execute ListSecrets to get a list of secrets.
	resp, err := client.ListSecrets(ctx, input)
	if err != nil {
		return nil, err
	}

	// Iterate through the list to find the secret with the desired name.
	for _, secret := range resp.SecretList {
		if aws.ToString(secret.Name) == secretName {
			return &secret, nil
		}
	}

	return nil, fmt.Errorf("secret with name '%s' not found", secretName)
}

func getSecretValue(ctx context.Context, client *secretsmanager.Client, secretID string) (string, error) {
	// Create an input for GetSecretValue.
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	// Execute GetSecretValue to retrieve the secret value.
	resp, err := client.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	// Check if the secret value is present in the response.
	if resp.SecretString == nil {
		return "", fmt.Errorf("secret value not found in the response")
	}

	return aws.ToString(resp.SecretString), nil
}

func checkJsonKeyString(jsonData []byte, key string) error {
	// Define a map to unmarshal the JSON data into.
	var data map[string]interface{}

	// Unmarshal the JSON data into the map.
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return fmt.Errorf("json invalid: %v", err)
	}

	// Check if the key exists in the map.
	val, exists := data[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	// Ensure key value type is string
	keyType := fmt.Sprintf("%T", val)
	if keyType != "string" && keyType != "float64" {
		return fmt.Errorf("key value type %s, expected string or float64", keyType)
	}

	return nil
}
