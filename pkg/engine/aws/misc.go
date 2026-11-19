package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func SplitResource(a arn.ARN) (resourceType string, identifier string) {
// Check earliest delimiter ('/' or ':')
slashIdx := strings.IndexByte(a.Resource, '/')
colonIdx := strings.IndexByte(a.Resource, ':')

sepIdx := -1
switch {
case slashIdx >= 0 && colonIdx >= 0:
sepIdx = min(slashIdx, colonIdx)
case slashIdx >= 0:
sepIdx = slashIdx
case colonIdx >= 0:
sepIdx = colonIdx
}

if sepIdx == -1 {
return "", a.Resource
}
return a.Resource[:sepIdx], a.Resource[sepIdx+1:]
}
