package gcp

import (
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
)

// gcpSecretsManagerSecret fetches secret object from GCP SecretsManager
func (e *GcpEngine) gcpSecretsManagerSecret(identifier ...string) (interface{}, error) {
	var (
		secretIdentifier    string
		projectID, secretID string
		version             = "latest"
	)

	switch len(identifier) {
	case 3:
		projectID = identifier[0]
		secretID = identifier[1]
		version = identifier[2]
		secretIdentifier = fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, version)
	case 2:
		projectID = identifier[0]
		secretID = identifier[1]
		secretIdentifier = fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, version)
	case 1:
		secretIdentifier = identifier[0]
	default:
		return nil, errors.New("invalid GCP SecretManager identifier, must be either <identifier> (path), <projectID, secretID> or <projectID, secretID, version>")
	}

	e.Logger.Info(`fetching GCP SecretsManager secret`, slog.String("identifier", secretIdentifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`gcpSecretsManagerSecret`, secretIdentifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := secretmanager.NewClient(e.Ctx)
		if err != nil {
			return "", fmt.Errorf("failed to create GCP SecretsManager client: %w", err)
		}
		defer client.Close()

		result, err := client.AccessSecretVersion(
			e.Ctx,
			&secretmanagerpb.AccessSecretVersionRequest{
				Name: secretIdentifier,
			},
		)
		if err != nil {
			return "", fmt.Errorf("failed to access secret version: %w", err)
		}

		crc32c := crc32.MakeTable(crc32.Castagnoli)
		checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
		if checksum != *result.Payload.DataCrc32C {
			return "", fmt.Errorf("failed to fetch GCP SecretsManager secret: data corruption detected, checksum mismatch")
		}

		return e.TransformToInterface(result)
	})
}

// gcpSecretsManagerSecretVersions fetches secret versions from GCP SecretsManager
func (e *GcpEngine) gcpSecretsManagerSecretVersions(identifier ...string) (interface{}, error) {
	var (
		secretIdentifier    string
		projectID, secretID string
		version             = "latest"
	)

	switch len(identifier) {
	case 2:
		projectID = identifier[0]
		secretID = identifier[1]
		secretIdentifier = fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, version)
	case 1:
		secretIdentifier = identifier[0]
	default:
		return nil, errors.New("invalid GCP SecretManager identifier, must be either <identifier> (path) or <projectID, secretID>")
	}

	e.Logger.Info(`fetching GCP SecretsManager secret versions`, slog.String("identifier", secretIdentifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`gcpSecretsManagerSecretVersions`, secretIdentifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := secretmanager.NewClient(e.Ctx)
		if err != nil {
			return "", fmt.Errorf("failed to create GCP SecretsManager client: %w", err)
		}
		defer client.Close()

		pager := client.ListSecretVersions(
			e.Ctx,
			&secretmanagerpb.ListSecretVersionsRequest{
				Parent: secretIdentifier,
			},
		)

		var results []*secretmanagerpb.SecretVersion
		for {
			version, err := pager.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed during iteration: %w", err)
			}

			results = append(results, version)
		}

		return e.TransformToInterface(results)
	})
}

// gcpSecretsManagerSecretList fetches list of secrets from GCP SecretsManager
func (e *GcpEngine) gcpSecretsManagerSecretList(identifier ...string) (interface{}, error) {
	var (
		secretIdentifier  string
		projectID, filter string
	)

	switch len(identifier) {
	case 2:
		filter = identifier[1]
		fallthrough
	case 1:
		if strings.HasPrefix(identifier[0], "projects/") {
			secretIdentifier = identifier[0]
		} else {
			secretIdentifier = fmt.Sprintf("projects/%s", projectID)
		}
	default:
		return nil, errors.New("invalid GCP SecretManager identifier, must be either <identifier> (path) or <projectID>")
	}

	e.Logger.Info(`fetching GCP SecretsManager secret list`, slog.String("identifier", secretIdentifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`gcpSecretsManagerSecretList`, secretIdentifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := secretmanager.NewClient(e.Ctx)
		if err != nil {
			return "", fmt.Errorf("failed to create GCP SecretsManager client: %w", err)
		}
		defer client.Close()

		pager := client.ListSecrets(
			e.Ctx,
			&secretmanagerpb.ListSecretsRequest{
				Parent: secretIdentifier,
				Filter: filter,
			},
		)

		var results []*secretmanagerpb.Secret
		for {
			version, err := pager.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed during iteration: %w", err)
			}

			results = append(results, version)
		}

		return e.TransformToInterface(results)
	})
}
