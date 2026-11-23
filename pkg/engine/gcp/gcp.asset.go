package gcp

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
)

// gcpAssetResource fetches a resource from Asset api
func (e *GcpEngine) gcpAssetResource(scope, assetType, identifier string) (interface{}, error) {
	e.Logger.Info(`fetching GCP asset resource`, slog.String("scope", scope), slog.String("assetType", assetType), slog.String("identifier", identifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`gcpAssetResource`, scope, assetType, identifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := asset.NewClient(e.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP Asset client: %w", err)
		}

		escapedID := escapeQueryValue(identifier)
		req := &assetpb.SearchAllResourcesRequest{
			Scope: scope,
			Query: fmt.Sprintf("name:%s", escapedID),
		}

		if assetType != "" {
			req.AssetTypes = []string{assetType}
		}

		pager := client.SearchAllResources(e.Ctx, req)
		result, err := pager.Next()
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch GCP Asset resource": %w`, err)
		}

		return e.TransformToInterface(result)
	})
}

// gcpAssetResource fetches a resource from Asset api
func (e *GcpEngine) gcpAssetQuery(scope string, assetTypes []string, query string) (interface{}, error) {
	e.Logger.Info(`fetching GCP asset resource by query`, slog.String("scope", scope), slog.Any("assetTypes", assetTypes), slog.String("query", query))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`gcpAssetQuery`, scope, strings.Join(assetTypes, ";"), query)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := asset.NewClient(e.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP Asset client: %w", err)
		}

		req := &assetpb.SearchAllResourcesRequest{
			Scope:      scope,
			AssetTypes: assetTypes,
			Query:      query,
		}

		var results []interface{}
		pager := client.SearchAllResources(e.Ctx, req)
		for {
			res, err := pager.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				log.Fatalf("error searching resources: %v", err)
			}

			results = append(results, res)
		}

		return e.TransformToInterface(results)
	})

}
