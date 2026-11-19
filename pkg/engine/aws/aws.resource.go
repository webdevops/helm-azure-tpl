package aws

import (
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
)

// awsResource fetches resource json from via resourceexporter2 by arn
func (e *AwsEngine) awsResource(arn string) (interface{}, error) {
	e.Logger.Info(`fetching AWS resource`, slog.String("arn", arn))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsResource`, arn)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		// 3. Query AWS Resource Explorer for the exact ARN
		client := resourceexplorer2.NewFromConfig(e.AwsConfig())
		out, err := client.Search(e.Ctx,
			&resourceexplorer2.SearchInput{
				QueryString: aws.String(fmt.Sprintf("arn:%s", arn)),
				MaxResults:  aws.Int32(1),
			},
		)
		if err != nil {
			return nil, fmt.Errorf(`unable to fetch AWS resource '%v': %w`, arn, err)
		}

		if len(out.Resources) == 0 {
			return nil, fmt.Errorf(`no AWS resource found by ARN '%v'`, arn)
		}

		return e.TransformToInterface(out)
	})
}
