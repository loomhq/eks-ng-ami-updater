package aws

import (
	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type testEks struct {
	OutputListClusters      *eks.ListClustersOutput
	OutputListNodegroups    *eks.ListNodegroupsOutput
	OutputDescribeNodegroup *eks.DescribeNodegroupOutput
}

type testEc2 struct {
	Output *ec2.DescribeRegionsOutput
}

type testSsm struct {
	OutputGetParameter *ssm.GetParameterOutput
}

func (t testEks) ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error) {
	output := t.OutputListClusters

	if t.OutputListClusters.NextToken != nil {
		if *t.OutputListClusters.NextToken == "111" {
			t.OutputListClusters.NextToken = awsLib.String("222")
		} else if *t.OutputListClusters.NextToken == "222" {
			t.OutputListClusters.NextToken = nil
			output.Clusters = []*string{t.OutputListClusters.Clusters[0]}
		}
	}

	return output, nil
}

func (t testEks) ListNodegroups(input *eks.ListNodegroupsInput) (*eks.ListNodegroupsOutput, error) {
	output := t.OutputListNodegroups

	if t.OutputListNodegroups.NextToken != nil {
		if *t.OutputListNodegroups.NextToken == "111" {
			t.OutputListNodegroups.NextToken = awsLib.String("222")
		} else if *t.OutputListNodegroups.NextToken == "222" {
			t.OutputListNodegroups.NextToken = nil
			output.Nodegroups = []*string{t.OutputListNodegroups.Nodegroups[0]}
		}
	}

	return output, nil
}

func (t testEks) DescribeNodegroup(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error) {
	output := t.OutputDescribeNodegroup

	return output, nil
}

func (t testEc2) DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	var output = t.Output

	return output, nil
}

func (t testSsm) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	var output = t.OutputGetParameter

	return output, nil
}
