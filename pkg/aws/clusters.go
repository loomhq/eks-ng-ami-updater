package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/rs/zerolog/log"
)

func GetClusters(region string, awsEks EKS, ctx context.Context) ([]string, error) {
	var clusters []string
	logWithContext := log.Ctx(ctx).With().Str("function", "GetClusters").Logger()

	logWithContext.Debug().Str("region", region).Msg("looking for clusters")

	input := &eks.ListClustersInput{}

	for {
		output, err := awsEks.ListClusters(input)
		if err != nil {
			return nil, err
		}
		for _, v := range output.Clusters {
			clusters = append(clusters, *v)
		}
		if output.NextToken != nil {
			logWithContext.Debug().Str("region", region).Strs("clusters", clusters).Str("token", *output.NextToken).Msg("ListClusters request exceed maxResults")
			input = &eks.ListClustersInput{
				NextToken: output.NextToken,
			}
		} else {
			break
		}
	}

	logWithContext.Debug().Str("region", region).Strs("clusters", clusters).Msg("clusters have been selected")

	return clusters, nil
}
