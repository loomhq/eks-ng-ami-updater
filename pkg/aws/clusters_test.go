package aws

import (
	"context"
	"fmt"
	"testing"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/stretchr/testify/assert"
)

func TestGetClusters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mockedOutput  eks.ListClustersOutput
		region        string
		expectedValue []string
		expectedError error
	}{
		{name: "full listing at once (without nextToken)",
			mockedOutput: eks.ListClustersOutput{
				Clusters: []*string{
					awsLib.String("cluster-1"),
					awsLib.String("cluster-2"),
				},
				NextToken: nil,
			},
			region:        "us-west-1",
			expectedValue: []string{"cluster-1", "cluster-2"},
			expectedError: nil,
		},
		{name: "exceed maxResults (nextToken is set)",
			mockedOutput: eks.ListClustersOutput{
				Clusters: []*string{
					awsLib.String("cluster-1"),
					awsLib.String("cluster-2"),
				},
				NextToken: awsLib.String("111"),
			},
			region:        "us-west-1",
			expectedValue: []string{"cluster-1", "cluster-2", "cluster-1"},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		fmt.Printf("test: %s\n", test.name)
		awsEks := testEks{OutputListClusters: &test.mockedOutput}

		output, err := GetClusters(test.region, awsEks, context.Background())

		assert.Equal(t, test.expectedValue, output)
		assert.Equal(t, test.expectedError, err)
	}
}
