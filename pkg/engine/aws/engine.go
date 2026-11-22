package aws

import (
	"log"
	textTemplate "text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/webdevops/helm-azure-tpl/pkg/engine/base"
)

type (
	AwsEngine struct {
		base.Engine
		*base.BaseEngine

		awsConfig *aws.Config
	}
)

func New(base *base.BaseEngine) *AwsEngine {
	ret := AwsEngine{
		BaseEngine: base,
	}
	ret.Cache = ret.Cache.Sub("aws")
	return &ret
}

func (e *AwsEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	funcMap := textTemplate.FuncMap{
		// Cloudcontrol
		`awsCloudControlResource`: e.awsCloudControlResource,

		// EC2
		`awsEc2Vpc`:    e.awsEc2Vpc,
		`awsEc2Subnet`: e.awsEc2Subnet,
		`awsEc2Ipam`:   e.awsEc2Ipam,

		// SecretsManager
		`awsSecretsManagerSecret`:         e.awsSecretsManagerSecret,
		`awsSecretsManagerSecretVersions`: e.awsSecretsManagerSecretVersions,
		`awsSecretsManagerSecretList`:     e.awsSecretsManagerSecretList,
	}

	return funcMap
}

// AwsConfig creates on-demand aws config and returns it
func (e *AwsEngine) AwsConfig() aws.Config {
	if e.awsConfig == nil {
		// Load default AWS configuration (~/.aws/config, env vars, or IAM role)
		cfg, err := config.LoadDefaultConfig(e.Ctx)
		if err != nil {
			log.Fatalf("unable to load SDK config: %v", err)
		}

		e.awsConfig = &cfg
	}

	return *e.awsConfig
}
