package aws

import (
	"context"
	"fmt"
	"testing"
	"time"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
)

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func TestIsLastAmiOldEnough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                        string
		mockedOutputGetParameterSsm ssm.GetParameterOutput
		mockedOutputGetParameterEc2 ec2.DescribeImagesOutput
		skipNewerThanDays           uint
		nodegroup                   NodeGroup
		ngAmiType                   string
		ngAmiVersion                string
		today                       time.Time
		expectedValue               bool
		expectedError               error
	}{
		{
			name:         "latest ami in aws is not newer than skipNewerThanDays (ami will be updated)",
			ngAmiType:    "BOTTLEROCKET_x86_64",
			ngAmiVersion: "1.24",
			mockedOutputGetParameterSsm: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					LastModifiedDate: toTimePtr(time.Date(2023, time.January, 30, 10, 0, 0, 0, time.UTC)),
					Value:            awsLib.String("ami-08a3df9f52daf9b5f"),
				},
			},
			mockedOutputGetParameterEc2: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageLocation: awsLib.String("amazon/bottlerocket-aws-k8s-1.24-x86_64-v1.14.0-9cd59298"),
					},
				},
			},
			skipNewerThanDays: 10,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.April, 31, 10, 0, 0, 0, time.UTC),
			expectedValue:     true,
			expectedError:     nil,
		},
		{
			name:         "latest ami in aws is newer than skipNewerThanDays (ami will not be updated)",
			ngAmiType:    "BOTTLEROCKET_x86_64",
			ngAmiVersion: "1.24",
			mockedOutputGetParameterSsm: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					LastModifiedDate: toTimePtr(time.Date(2023, time.January, 30, 10, 0, 0, 0, time.UTC)),
					Value:            awsLib.String("ami-08a3df9f52daf9b5f"),
				},
			},
			mockedOutputGetParameterEc2: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageLocation: awsLib.String("amazon/bottlerocket-aws-k8s-1.24-x86_64-v1.14.0-9cd59298"),
					},
				},
			},
			skipNewerThanDays: 10,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.February, 1, 10, 0, 0, 0, time.UTC),
			expectedValue:     false,
			expectedError:     nil,
		},
		{
			name:         "one day and 1 minute diff",
			ngAmiType:    "BOTTLEROCKET_x86_64",
			ngAmiVersion: "1.24",
			mockedOutputGetParameterSsm: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					LastModifiedDate: toTimePtr(time.Date(2023, time.January, 15, 9, 59, 0, 0, time.UTC)),
					Value:            awsLib.String("ami-08a3df9f52daf9b5f"),
				},
			},
			mockedOutputGetParameterEc2: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageLocation: awsLib.String("amazon/bottlerocket-aws-k8s-1.24-x86_64-v1.14.0-9cd59298"),
					},
				},
			},
			skipNewerThanDays: 1,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.January, 16, 10, 0, 0, 0, time.UTC),
			expectedValue:     true,
			expectedError:     nil,
		},
		{
			name:         "unrecognize ami type",
			ngAmiType:    "SOMETHINGNEW_x86_64",
			ngAmiVersion: "1.24",
			mockedOutputGetParameterSsm: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					LastModifiedDate: toTimePtr(time.Date(2023, time.January, 30, 10, 0, 0, 0, time.UTC)),
					Value:            awsLib.String("ami-08a3df9f52daf9b5f"),
				},
			},
			mockedOutputGetParameterEc2: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageLocation: awsLib.String("amazon/bottlerocket-aws-k8s-1.24-x86_64-v1.14.0-9cd59298"),
					},
				},
			},
			skipNewerThanDays: 10,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.February, 1, 10, 0, 0, 0, time.UTC),
			expectedValue:     false,
			expectedError:     fmt.Errorf("region: , cluster: , nodegroup:  : %w", fmt.Errorf("nodegroup's ami type (SOMETHINGNEW_x86_64) is not recognize")),
		},
		{
			name:         "AL2 ami type",
			ngAmiType:    "AL2_x86_64",
			ngAmiVersion: "1.28.",
			mockedOutputGetParameterSsm: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					LastModifiedDate: toTimePtr(time.Date(2023, time.January, 30, 10, 0, 0, 0, time.UTC)),
					Value:            awsLib.String("ami-08a3df9f52daf9b5f"),
				},
			},
			mockedOutputGetParameterEc2: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageLocation: awsLib.String("amazon/amazon-eks-node-1.28-v20240202"),
					},
				},
			},
			skipNewerThanDays: 10,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.February, 1, 10, 0, 0, 0, time.UTC),
			expectedValue:     false,
			expectedError:     nil,
		},
	}

	for _, test := range tests {
		fmt.Printf("test: %s\n", test.name)
		awsSsm := testSsm{
			OutputGetParameter: &test.mockedOutputGetParameterSsm,
		}
		awsEc2 := testEc2{
			OutputImages: &test.mockedOutputGetParameterEc2,
		}

		output, err := IsLastAmiOldEnough(test.skipNewerThanDays, test.nodegroup, test.today, test.ngAmiType, test.ngAmiVersion, awsSsm, awsEc2, context.Background())

		assert.Equal(t, test.expectedValue, output)
		assert.Equal(t, test.expectedError, err)
	}
}
