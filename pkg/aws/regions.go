package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rs/zerolog/log"
)

func GetRegionsToCheck(regionsVar []string, awsEc2 Ec2, ctx context.Context) ([]string, error) {
	var regions []string
	logWithContext := log.Ctx(ctx).With().Str("function", "GetRegionsToCheck").Logger()

	if len(regionsVar) == 0 {
		logWithContext.Debug().Msg("looking for regions")

		input := &ec2.DescribeRegionsInput{AllRegions: aws.Bool(true)}
		result, err := awsEc2.DescribeRegions(input)
		if err != nil {
			return nil, err
		}

		for _, v := range result.Regions {
			regions = append(regions, *v.RegionName)
		}

		logWithContext.Debug().Strs("regions", regions).Msg("regions have been selected")
	} else {
		regions = regionsVar
	}

	return regions, nil
}
