package aws

import (
	"context"
	"fmt"
	"time"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/rs/zerolog/log"
)

func GetLatestAmi(amiType, amiVersion, region string, awsSsm SSM, ctx context.Context) (time.Time, error) {
	var ssmPath string
	logWithContext := log.Ctx(ctx).With().Str("function", "GetLatestAmi").Logger()

	switch amiType {
	// https://docs.aws.amazon.com/eks/latest/userguide/eks-optimized-ami-bottlerocket.html
	case "BOTTLEROCKET_x86_64":
		ssmPath = "/aws/service/bottlerocket/aws-k8s-" + amiVersion + "/x86_64/latest/image_id"
	case "BOTTLEROCKET_x86_64_NVIDIA":
		ssmPath = "/aws/service/bottlerocket/aws-k8s-" + amiVersion + "-nvidia/x86_64/latest/image_id"
	case "BOTTLEROCKET_ARM_64":
		ssmPath = "/aws/service/bottlerocket/aws-k8s-" + amiVersion + "/arm64/latest/image_id"
	case "BOTTLEROCKET_ARM_64_NVIDIA":
		ssmPath = "/aws/service/bottlerocket/aws-k8s-" + amiVersion + "-nvidia/arm64/latest/image_id"
	// https://docs.aws.amazon.com/eks/latest/userguide/eks-optimized-ami.html
	case "AL2_x86_64":
		ssmPath = "/aws/service/eks/optimized-ami/" + amiVersion + "/amazon-linux-2/recommended/image_id"
	case "AL2_x86_64_GPU":
		ssmPath = "/aws/service/eks/optimized-ami/" + amiVersion + "/amazon-linux-2-gpu/recommended/image_id"
	case "AL2_ARM_64":
		ssmPath = "/aws/service/eks/optimized-ami/" + amiVersion + "/amazon-linux-2-arm64/recommended/image_id"
	// https://docs.aws.amazon.com/eks/latest/userguide/retrieve-windows-ami-id.html
	case "WINDOWS_CORE_2019_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2019-English-Core-EKS_Optimized-" + amiVersion + "/image_id"
	case "WINDOWS_FULL_2019_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2019-English-Full-EKS_Optimized-" + amiVersion + "/image_id"
	case "WINDOWS_CORE_2022_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2022-English-Core-EKS_Optimized-" + amiVersion + "/image_id"
	case "WINDOWS_FULL_2022_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2022-English-Full-EKS_Optimized-" + amiVersion + "/image_id"
	default:
		return time.Time{}, fmt.Errorf("nodegroup's ami type (%s) is not recognize", amiType)
	}

	output, err := awsSsm.GetParameter(&ssm.GetParameterInput{
		Name:           &ssmPath,
		WithDecryption: new(bool),
	})
	if err != nil {
		return time.Time{}, err
	}

	logWithContext.Debug().Time("LastModifiedDate", *output.Parameter.LastModifiedDate).
		Str("region", region).Str("amiVersion", amiVersion).Str("amiType", amiType).Msg("ami LastModifiedDate is found")

	return *output.Parameter.LastModifiedDate, nil
}

func IsLastAmiOldEnough(skipNewerThan uint, nodegroup NodeGroup, today time.Time, ngAmiType, ngAmiVersion string, awsSsm SSM, ctx context.Context) (bool, error) {
	logWithContext := log.Ctx(ctx).With().Str("function", "IsLastAmiOldEnough").Logger()

	amiLastModifiedDate, err := GetLatestAmi(ngAmiType, ngAmiVersion, nodegroup.Region, awsSsm, ctx)
	if err != nil {
		return false, fmt.Errorf("region: %s, cluster: %s, nodegroup: %s : %w", nodegroup.Region, nodegroup.ClusterName, nodegroup.ClusterName, err)
	}

	hoursToSkip := -24 * time.Duration(skipNewerThan) * time.Hour
	criticalDay := today.Add(hoursToSkip).UTC()

	logWithContext.Debug().Time("criticalDay", criticalDay).Uint("skipNewerThan", skipNewerThan).
		Str("region", nodegroup.Region).Str("nodegroup", nodegroup.NodegroupName).Str("cluster", nodegroup.ClusterName).Msg("ami criticalDay is calculated")

	if amiLastModifiedDate.Before(criticalDay) {
		return true, nil
	}

	return false, nil
}

func AmiUpdate(region, cluster, nodegroup string, dryrun bool, ctx context.Context) error {
	logWithContext := log.Ctx(ctx).With().Str("function", "AmiUpdate").Logger()

	if dryrun {
		log.Info().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Msg("drying is true. exiting")

		return nil
	}

	svcEks, err := EksClientSetup(region)
	if err != nil {
		logWithContext.Debug().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Err(err).Msg("Error")

		return err
	}

	log.Info().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Msg("starting ami update")

	_, err = svcEks.UpdateNodegroupVersion(&eks.UpdateNodegroupVersionInput{
		ClusterName:   awsLib.String(cluster),
		NodegroupName: awsLib.String(nodegroup),
	})
	if err != nil {
		logWithContext.Debug().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Err(err).Msg("Error")

		return err
	}
	logWithContext.Debug().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Msg("started properly")

	logWithContext.Debug().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Msg("waiting for finish")
	err = svcEks.WaitUntilNodegroupActive(&eks.DescribeNodegroupInput{
		ClusterName:   awsLib.String(cluster),
		NodegroupName: awsLib.String(nodegroup),
	})
	if err != nil {
		logWithContext.Debug().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Err(err).Msg("Error")

		return err
	}

	log.Info().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Msg("finished ami update properly")

	return nil
}
