package aws

import (
	"context"
	"fmt"
	"testing"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/stretchr/testify/assert"
)

func TestGetNodegroupsFromCluster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mockedOutput  eks.ListNodegroupsOutput
		clusterName   string
		region        string
		expectedValue []string
		expectedError error
	}{
		{name: "full listing at once (without nextToken)",
			mockedOutput: eks.ListNodegroupsOutput{
				Nodegroups: []*string{
					awsLib.String("nodegroup-1"),
					awsLib.String("nodegroup-2"),
				},
			},
			clusterName:   "cluster-1",
			region:        "us-west-1",
			expectedValue: []string{"nodegroup-1", "nodegroup-2"},
			expectedError: nil,
		},
		{name: "exceed maxResults (nextToken is set)",
			mockedOutput: eks.ListNodegroupsOutput{
				NextToken: awsLib.String("111"),
				Nodegroups: []*string{
					awsLib.String("nodegroup-1"),
					awsLib.String("nodegroup-2"),
				},
			},
			clusterName:   "cluster-1",
			region:        "us-west-1",
			expectedValue: []string{"nodegroup-1", "nodegroup-2", "nodegroup-1"},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		test := test

		fmt.Printf("test: %s\n", test.name)
		awsEks := testEks{OutputListNodegroups: &test.mockedOutput}

		output, err := GetNodegroupsFromCluster(test.clusterName, test.region, awsEks, context.Background())

		assert.Equal(t, test.expectedValue, output)
		assert.Equal(t, test.expectedError, err)
	}
}

func TestGetNodegroupTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tag           string
		nodegroupTags map[string]*string
		expectedValue bool
		expectedError error
	}{
		{name: "tag is found in nodegroup Tags",
			tag: "env:staging",
			nodegroupTags: map[string]*string{
				"tag1": awsLib.String("value1"),
				"env":  awsLib.String("staging"),
				"tag3": awsLib.String("value3"),
			},
			expectedValue: true,
			expectedError: nil,
		},
		{name: "tag is not found in nodegroup Tags",
			tag: "env:production",
			nodegroupTags: map[string]*string{
				"tag1": awsLib.String("value1"),
				"env":  awsLib.String("staging"),
				"tag3": awsLib.String("value3"),
			},
			expectedValue: false,
			expectedError: nil,
		},
		{name: "nodegroup tags list is empty",
			tag:           "env:staging",
			nodegroupTags: map[string]*string{},
			expectedValue: false,
			expectedError: nil,
		},
		{name: "only key is provided for tag #1",
			tag: "env",
			nodegroupTags: map[string]*string{
				"tag1": awsLib.String("value1"),
				"env":  awsLib.String("staging"),
				"tag3": awsLib.String("value3"),
			},
			expectedValue: false,
			expectedError: nil,
		},
		{name: "only key is provided for tag #2",
			tag: "env:",
			nodegroupTags: map[string]*string{
				"tag1": awsLib.String("value1"),
				"env":  awsLib.String(""),
				"tag3": awsLib.String("value3"),
			},
			expectedValue: true,
			expectedError: nil,
		},
		{name: "only value is provided for tag #1",
			tag: ":staging",
			nodegroupTags: map[string]*string{
				"tag1": awsLib.String("value1"),
				"env":  awsLib.String("staging"),
				"tag3": awsLib.String("value3"),
			},
			expectedValue: false,
			expectedError: nil,
		},
		{name: "only value is provided for tag #2",
			tag: ":staging",
			nodegroupTags: map[string]*string{
				"tag1": awsLib.String("value1"),
				"":     awsLib.String("staging"),
				"tag3": awsLib.String("value3"),
			},
			expectedValue: true,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		fmt.Printf("test: %s\n", test.name)

		output, err := HasNodegroupTag(test.tag, test.nodegroupTags, context.Background())

		assert.Equal(t, test.expectedValue, output)
		assert.Equal(t, test.expectedError, err)
	}
}
