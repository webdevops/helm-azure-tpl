package hashicorp

import (
	"log"
	"log/slog"
)

// HashicorpVaultSecret fetches an secret from Hashicorp vault
func (e *HashicorpEngine) HashicorpVaultSecret(secretPath string) (interface{}, error) {
	return e.HashicorpVaultGet("secret", secretPath)
}

// hashicorpVaultGet fetches an object from Hashicorp vault
func (e *HashicorpEngine) HashicorpVaultGet(mountPath, secretPath string) (interface{}, error) {
	e.Logger.Info(`fetching Hashicorp Vault item`, slog.String("mountPath", mountPath), slog.String("secretPath", secretPath))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`hashicorpVaultGet`, mountPath, secretPath)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client := e.VaultClient()
		secret, err := client.KVv2(mountPath).Get(e.Ctx, secretPath)
		if err != nil {
			log.Fatalf("unable to fetch Hashicorp Vault item: %v", err)
		}
		return e.TransformToInterface(secret)
	})
}
