package aws

import (
	"context"
	"fmt"
	"testing"
	"time"

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
				},
			},
			skipNewerThanDays: 1,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.January, 16, 10, 0, 0, 0, time.UTC),
			expectedValue:     true,
			expectedError:     nil,
		},
		{
			name:         "unrecognize ami type)",
			ngAmiType:    "SOMETHINGNEW_x86_64",
			ngAmiVersion: "1.24",
			mockedOutputGetParameterSsm: ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					LastModifiedDate: toTimePtr(time.Date(2023, time.January, 30, 10, 0, 0, 0, time.UTC)),
				},
			},
			skipNewerThanDays: 10,
			nodegroup:         NodeGroup{},
			today:             time.Date(2023, time.February, 1, 10, 0, 0, 0, time.UTC),
			expectedValue:     false,
			expectedError:     fmt.Errorf("region: , cluster: , nodegroup:  : %w", fmt.Errorf("nodegroup's ami type (SOMETHINGNEW_x86_64) is not recognize")),
		},
	}

	for _, test := range tests {
		fmt.Printf("test: %s\n", test.name)
		awsSsm := testSsm{
			OutputGetParameter: &test.mockedOutputGetParameterSsm,
		}

		output, err := IsLastAmiOldEnough(test.skipNewerThanDays, test.nodegroup, test.today, test.ngAmiType, test.ngAmiVersion, awsSsm, context.Background())

		assert.Equal(t, test.expectedValue, output)
		assert.Equal(t, test.expectedError, err)
	}
}
