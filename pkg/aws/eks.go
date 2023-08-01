package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/eks"
)

type EKS interface {
	ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error)
	ListNodegroups(input *eks.ListNodegroupsInput) (*eks.ListNodegroupsOutput, error)
	DescribeNodegroup(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error)
}

type RealEks struct {
	Svc *eks.EKS
}

func (t RealEks) ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error) {
	result, err := t.Svc.ListClusters(input)
	if err != nil {
		return nil, fmt.Errorf("error listing clusters: %w", err)
	}

	return result, nil
}

func (t RealEks) ListNodegroups(input *eks.ListNodegroupsInput) (*eks.ListNodegroupsOutput, error) {
	result, err := t.Svc.ListNodegroups(input)
	if err != nil {
		return nil, fmt.Errorf("error listing nodegroups: %w", err)
	}

	return result, nil
}

func (t RealEks) DescribeNodegroup(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error) {
	result, err := t.Svc.DescribeNodegroup(input)
	if err != nil {
		return nil, fmt.Errorf("error describing nodegroup: %w", err)
	}

	return result, nil
}
