package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type Ec2 interface {
	DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
}

type RealEc2 struct {
	Svc *ec2.EC2
}

func (t RealEc2) DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	result, err := t.Svc.DescribeRegions(input)
	if err != nil {
		return nil, fmt.Errorf("error describing regions: %w", err)
	}

	return result, nil
}
