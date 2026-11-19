package aws

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
)

// awsCloudControlResource fetches resource json from via resourceexporter2 by arn
func (e *AwsEngine) awsCloudControlResource(typeName, identifier string) (interface{}, error) {
	e.Logger.Info(`fetching AWS cloudcontrol resource`, slog.String("typeName", typeName), slog.String("identifier", identifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsCloudcontrolResource`, typeName, identifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client := cloudcontrol.NewFromConfig(e.AwsConfig())
		out, err := client.GetResource(
			e.Ctx,
			&cloudcontrol.GetResourceInput{
				TypeName:   aws.String(typeName),
				Identifier: aws.String(identifier),
			},
		)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch AWS cloudcontrol resource '%v': %w`, identifier, err)
		}

		var resourceProperties interface{}
		props := aws.ToString(out.ResourceDescription.Properties)
		err = json.Unmarshal([]byte(props), &resourceProperties)
		if err != nil {
			return nil, fmt.Errorf(`unable to parse AWS cloudcontrol resource properties '%v': %w`, identifier, err)
		}

		return e.TransformToInterface(resourceProperties)
	})
}
