package aws

import (
	"context"
	"fmt"
	"testing"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestGetRegionsToCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mockedOutput  ec2.DescribeRegionsOutput
		regionsVar    []string
		expectedValue []string
		expectedError error
	}{
		{name: "regionsVar cli parameter is set #1",
			mockedOutput: ec2.DescribeRegionsOutput{
				Regions: []*ec2.Region{
					{RegionName: awsLib.String("us-west-1")},
					{RegionName: awsLib.String("us-west-2")},
					{RegionName: awsLib.String("us-west-3")},
				},
			},
			regionsVar:    []string{"eu-west-1", "eu-west-2", "eu-west-3"},
			expectedValue: []string{"eu-west-1", "eu-west-2", "eu-west-3"},
			expectedError: nil,
		},
		{name: "regionsVar cli parameter is set #2",
			mockedOutput:  ec2.DescribeRegionsOutput{},
			regionsVar:    []string{"eu-west-1"},
			expectedValue: []string{"eu-west-1"},
			expectedError: nil,
		},
		{name: "regionsVar cli parameter is not set #1",
			mockedOutput: ec2.DescribeRegionsOutput{
				Regions: []*ec2.Region{
					{RegionName: awsLib.String("us-west-1")},
				},
			},
			regionsVar:    []string{},
			expectedValue: []string{"us-west-1"},
			expectedError: nil,
		},
		{name: "regionsVar cli parameter is not set #2",
			mockedOutput: ec2.DescribeRegionsOutput{
				Regions: []*ec2.Region{
					{RegionName: awsLib.String("us-west-1")},
					{RegionName: awsLib.String("us-west-2")},
					{RegionName: awsLib.String("us-west-3")},
				},
			},
			regionsVar:    []string{},
			expectedValue: []string{"us-west-1", "us-west-2", "us-west-3"},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		test := test

		fmt.Printf("test: %s\n", test.name)
		awsEc2 := testEc2{Output: &test.mockedOutput}
		regionsVar := test.regionsVar

		output, err := GetRegionsToCheck(regionsVar, awsEc2, context.Background())

		assert.Equal(t, test.expectedValue, output)
		assert.Equal(t, test.expectedError, err)
	}
}
