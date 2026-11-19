package aws

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// awsSecretsManagerSecret fetches secret object from AWS SecretsManager
func (e *AwsEngine) awsSecretsManagerSecret(secretName string, opts ...string) (interface{}, error) {
	version := "AWSCURRENT"
	if len(opts) == 1 {
		version = opts[0]
	}

	e.Logger.Info(`fetching AWS SecretsManager secret`, slog.String("secret", secretName), slog.String("version", version))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsSecretsManagerSecret`, secretName, strings.Join(opts, ";"))
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client := secretsmanager.NewFromConfig(e.AwsConfig())

		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String(version),
		}

		result, err := client.GetSecretValue(e.Ctx, input)
		if err != nil {
			return nil, fmt.Errorf(`failed to retrieve AWS SecretsManager secret "%s": %w`, secretName, err)
		}

		return e.TransformToInterface(result)
	})
}

// awsSecretsManagerSecretVersions fetches secret versions from AWS SecretsManager
func (e *AwsEngine) awsSecretsManagerSecretVersions(secretName string, count int) (interface{}, error) {
	e.Logger.Info(`fetching AWS SecretsManager secret versions`, slog.String("secret", secretName), slog.Int("count", count))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsSecretsManagerSecretVersions`, secretName, strconv.Itoa(count))
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client := secretsmanager.NewFromConfig(e.AwsConfig())

		input := &secretsmanager.ListSecretVersionIdsInput{
			SecretId: aws.String(secretName),
		}
		paginator := secretsmanager.NewListSecretVersionIdsPaginator(client, input)

		var results []types.SecretVersionsListEntry
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(e.Ctx)
			if err != nil {
				return nil, fmt.Errorf(`failed to retrieve AWS SecretsManager secret list: %w`, err)
			}
			results = append(results, page.Versions...)
		}

		return e.TransformToInterface(results)
	})
}

// awsSecretsManagerSecretList fetches secrets from AWS SecretsManager
func (e *AwsEngine) awsSecretsManagerSecretList() (interface{}, error) {
	e.Logger.Info(`fetching AWS SecretsManager secret list`)

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsSecretsManagerSecretList`)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client := secretsmanager.NewFromConfig(e.AwsConfig())

		input := &secretsmanager.ListSecretsInput{}
		paginator := secretsmanager.NewListSecretsPaginator(client, input)

		var results []types.SecretListEntry
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(e.Ctx)
			if err != nil {
				return nil, fmt.Errorf(`failed to retrieve AWS SecretsManager secret list: %w`, err)
			}
			results = append(results, page.SecretList...)
		}

		return e.TransformToInterface(results)
	})
}
