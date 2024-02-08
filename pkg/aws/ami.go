package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/rs/zerolog/log"
)

func GetLatestAmiWithinSsm(amiType, amiVersion, region string, awsSsm SSM, awsEc2 Ec2, ctx context.Context) (time.Time, *string, error) {
	var ssmPath string

	ssmBootlerocketPathPrefix := "/aws/service/bottlerocket/aws-k8s-"
	ssmOptimizedPathPrefix := "/aws/service/eks/optimized-ami/"
	ssmPathSuffix := "/image_id"

	switch amiType {
	// https://docs.aws.amazon.com/eks/latest/userguide/eks-optimized-ami-bottlerocket.html
	case "BOTTLEROCKET_x86_64":
		ssmPath = ssmBootlerocketPathPrefix + amiVersion + "/x86_64/latest" + ssmPathSuffix
	case "BOTTLEROCKET_x86_64_NVIDIA":
		ssmPath = ssmBootlerocketPathPrefix + amiVersion + "-nvidia/x86_64/latest" + ssmPathSuffix
	case "BOTTLEROCKET_ARM_64":
		ssmPath = ssmBootlerocketPathPrefix + amiVersion + "/arm64/latest" + ssmPathSuffix
	case "BOTTLEROCKET_ARM_64_NVIDIA":
		ssmPath = ssmBootlerocketPathPrefix + amiVersion + "-nvidia/arm64/latest" + ssmPathSuffix
	// https://docs.aws.amazon.com/eks/latest/userguide/eks-optimized-ami.html
	case "AL2_x86_64":
		ssmPath = ssmOptimizedPathPrefix + amiVersion + "/amazon-linux-2/recommended" + ssmPathSuffix
	case "AL2_x86_64_GPU":
		ssmPath = ssmOptimizedPathPrefix + amiVersion + "/amazon-linux-2-gpu/recommended" + ssmPathSuffix
	case "AL2_ARM_64":
		ssmPath = ssmOptimizedPathPrefix + amiVersion + "/amazon-linux-2-arm64/recommended" + ssmPathSuffix
	// https://docs.aws.amazon.com/eks/latest/userguide/retrieve-windows-ami-id.html
	case "WINDOWS_CORE_2019_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2019-English-Core-EKS_Optimized-" + amiVersion + ssmPathSuffix
	case "WINDOWS_FULL_2019_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2019-English-Full-EKS_Optimized-" + amiVersion + ssmPathSuffix
	case "WINDOWS_CORE_2022_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2022-English-Core-EKS_Optimized-" + amiVersion + ssmPathSuffix
	case "WINDOWS_FULL_2022_x86_64":
		ssmPath = "/aws/service/ami-windows-latest/Windows_Server-2022-English-Full-EKS_Optimized-" + amiVersion + ssmPathSuffix
	default:
		return time.Time{}, nil, fmt.Errorf("nodegroup's ami type (%s) is not recognize", amiType)
	}

	output, err := awsSsm.GetParameter(&ssm.GetParameterInput{
		Name:           &ssmPath,
		WithDecryption: new(bool),
	})
	if err != nil {
		return time.Time{}, nil, err
	}

	awsLatestAmiImageLocation, err := GetLatestAmiWithinEc2(*output.Parameter.Value, awsEc2, ctx)
	if err != nil {
		return time.Time{}, nil, err
	}

	return *output.Parameter.LastModifiedDate, awsLatestAmiImageLocation, nil
}

func GetLatestAmiWithinEc2(amiVersion string, awsEc2 Ec2, ctx context.Context) (*string, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []*string{
			awsLib.String(amiVersion),
		},
		Owners: []*string{
			awsLib.String("amazon"),
		},
	}

	result, err := awsEc2.DescribeImages(input)
	if err != nil {
		return nil, err
	}

	return result.Images[0].ImageLocation, nil
}

func IsTheSameAmiVersion(nodegroup NodeGroup, ngAmiType, ngAmiVersion, ngAmiReleaseVersion string, awsSsm SSM, awsEc2 Ec2, ctx context.Context) (bool, error) {
	var awsLatestAmiReleaseVersion string

	logWithContext := log.Ctx(ctx).With().Str("function", "IsTheSameAmiVersion").Logger()

	_, awsLatestAmiImageLocation, err := GetLatestAmiWithinSsm(ngAmiType, ngAmiVersion, nodegroup.Region, awsSsm, awsEc2, ctx)
	if err != nil {
		return false, fmt.Errorf("region: %s, cluster: %s, nodegroup: %s : %w", nodegroup.Region, nodegroup.ClusterName, nodegroup.ClusterName, err)
	}

	awsLatestAmiImageLocationSplites := strings.Split(*awsLatestAmiImageLocation, "-")
	last := len(awsLatestAmiImageLocationSplites)
	switch strings.Split(ngAmiType, "_")[0] {
	case "BOTTLEROCKET":
		awsLatestAmiReleaseVersion = awsLatestAmiImageLocationSplites[last-2] + "-" + awsLatestAmiImageLocationSplites[last-1]
		ngAmiReleaseVersion = "v" + ngAmiReleaseVersion
	case "AL2", "WINDOWS":
		awsLatestAmiReleaseVersion = awsLatestAmiImageLocationSplites[last-1]
		ngAmiReleaseVersion = "v" + strings.Split(ngAmiReleaseVersion, "-")[1]
	default:
		return true, fmt.Errorf("nodegroup's ami type (%s) is not recognize", ngAmiType)
	}

	isTheSameAmiVersion := true
	if awsLatestAmiReleaseVersion != ngAmiReleaseVersion {
		isTheSameAmiVersion = false
	}

	logWithContext.Debug().Str("ngAmiReleaseVersion", ngAmiReleaseVersion).Str("awsLatestAmiReleaseVersion", awsLatestAmiReleaseVersion).
		Str("region", nodegroup.Region).Str("nodegroup", nodegroup.NodegroupName).Str("cluster", nodegroup.ClusterName).Msg("nodegroup and aws latest ami versions are compared")

	return isTheSameAmiVersion, nil
}

func IsLastAmiOldEnough(skipNewerThan uint, nodegroup NodeGroup, today time.Time, ngAmiType, ngAmiVersion string, awsSsm SSM, awsEc2 Ec2, ctx context.Context) (bool, error) {
	logWithContext := log.Ctx(ctx).With().Str("function", "IsLastAmiOldEnough").Logger()

	amiLastModifiedDate, _, err := GetLatestAmiWithinSsm(ngAmiType, ngAmiVersion, nodegroup.Region, awsSsm, awsEc2, ctx)
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
