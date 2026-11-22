package aws

import (
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// awsEc2Vpc fetches vpc by arn
func (e *AwsEngine) awsEc2Vpc(identifier string) (interface{}, error) {
	e.Logger.Info(`fetching AWS EC2 VPC resource`, slog.String("identifier", identifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsEc2Vpc`, identifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		ec2Client := ec2.NewFromConfig(e.AwsConfig())

		resource, err := ec2Client.DescribeVpcs(e.Ctx, &ec2.DescribeVpcsInput{
			VpcIds: []string{identifier},
		})
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 VPC resource "%s": %w`, identifier, err)
		}

		if len(resource.Vpcs) == 0 {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 VPC resource "%s": no resource found`, identifier)
		}
		if len(resource.Vpcs) > 1 {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 VPC resource "%s": more then one resource found`, identifier)
		}

		return e.TransformToInterface(resource.Vpcs[0])
	})
}

// awsEc2Subnet fetches subnet by arn
func (e *AwsEngine) awsEc2Subnet(identifier string) (interface{}, error) {
	e.Logger.Info(`fetching AWS EC2 Subnet resource`, slog.String("identifier", identifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsEc2Subnets`, identifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		ec2Client := ec2.NewFromConfig(e.AwsConfig())

		resource, err := ec2Client.DescribeSubnets(e.Ctx, &ec2.DescribeSubnetsInput{
			SubnetIds: []string{identifier},
		})
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 Subnet resource "%s": %w`, identifier, err)
		}

		if len(resource.Subnets) == 0 {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 Subnet resource "%s": no resource found`, identifier)
		}
		if len(resource.Subnets) > 1 {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 Subnet resource "%s": more then one resource found`, identifier)
		}

		return e.TransformToInterface(resource.Subnets[0])
	})
}

// awsEc2Ipam fetches vpc by arn
func (e *AwsEngine) awsEc2Ipam(identifier string) (interface{}, error) {
	e.Logger.Info(`fetching AWS EC2 IPAM resource`, slog.String("identifier", identifier))

	if val, enabled := e.LintResult(); enabled {
		return val, nil
	}

	cacheKey := e.GenerateCacheKey(`awsEc2Ipam`, identifier)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		ec2Client := ec2.NewFromConfig(e.AwsConfig())

		resource, err := ec2Client.DescribeIpamPools(e.Ctx, &ec2.DescribeIpamPoolsInput{
			IpamPoolIds: []string{identifier},
		})
		if err != nil {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 VPC resource "%s": %w`, identifier, err)
		}

		if len(resource.IpamPools) == 0 {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 IPAM resource "%s": no resource found`, identifier)
		}
		if len(resource.IpamPools) > 1 {
			return nil, fmt.Errorf(`failed to fetch AWS EC2 IPAM resource "%s": more then one resource found`, identifier)
		}

		return e.TransformToInterface(resource.IpamPools[0])
	})
}
