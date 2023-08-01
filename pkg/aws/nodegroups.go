package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/rs/zerolog/log"
)

type NodeGroup struct {
	Region        string
	ClusterName   string
	NodegroupName string
}

func GetNodegroupsFromCluster(clusterName, region string, awsEks EKS, ctx context.Context) ([]string, error) {
	var nodegroups []string
	logWithContext := log.Ctx(ctx).With().Str("function", "GetNodegroupsFromCluster").Logger()

	logWithContext.Debug().Str("region", region).Str("cluster", clusterName).Msg("looking for nodegroups")
	input := &eks.ListNodegroupsInput{
		ClusterName: &clusterName,
	}

	for {
		output, err := awsEks.ListNodegroups(input)
		if err != nil {
			return nil, err
		}
		for _, v := range output.Nodegroups {
			nodegroups = append(nodegroups, *v)
		}

		if output.NextToken != nil {
			logWithContext.Debug().Str("region", region).Str("cluster", clusterName).Str("token", *output.NextToken).Msg("ListNodegroups request exceed maxResults")
			input = &eks.ListNodegroupsInput{
				NextToken: output.NextToken,
			}
		} else {
			break
		}
	}

	logWithContext.Debug().Str("region", region).Str("cluster", clusterName).Strs("nodegroup", nodegroups).Msg("nodegroups have been selected")

	return nodegroups, nil
}

func GetNodegroupsFromRegion(region string, ctx context.Context) ([]NodeGroup, error) {
	var nodeGroupFromCluster []NodeGroup
	logWithContext := log.Ctx(ctx).With().Str("function", "GetNodegroupsFromRegion").Logger()

	svcEks, err := EksClientSetup(region)
	if err != nil {
		return nil, err
	}
	awsEks := RealEks{Svc: svcEks}

	clusters, err := GetClusters(region, awsEks, ctx)
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		nodegroups, err := GetNodegroupsFromCluster(cluster, region, awsEks, ctx)
		if err != nil {
			return nil, err
		}
		for _, nodegroup := range nodegroups {
			nodeGroupFromCluster = append(nodeGroupFromCluster, NodeGroup{Region: region, ClusterName: cluster, NodegroupName: nodegroup})
			logWithContext.Debug().Str("region", region).Str("cluster", cluster).Str("nodegroup", nodegroup).Msg("add nodegroup to the ami upgrade nodegroups checking list")
		}
	}

	return nodeGroupFromCluster, nil
}

func GetNodegroupDescription(nodegroup NodeGroup, awsEks EKS, ctx context.Context) (eks.DescribeNodegroupOutput, error) {
	logWithContext := log.Ctx(ctx).With().Str("function", "GetNodegroupDescription").Logger()

	output, err := awsEks.DescribeNodegroup(&eks.DescribeNodegroupInput{
		ClusterName:   &nodegroup.ClusterName,
		NodegroupName: &nodegroup.NodegroupName,
	})
	if err != nil {
		return eks.DescribeNodegroupOutput{}, err
	}

	logWithContext.Debug().Str("AmiType", *output.Nodegroup.AmiType).Str("AmiVersion", *output.Nodegroup.Version).
		Str("region", nodegroup.Region).Str("cluster", nodegroup.ClusterName).Str("nodegroup", nodegroup.NodegroupName).Msg("actual nodegroup AMI Id")

	for k, v := range output.Nodegroup.Tags {
		logWithContext.Debug().Str("tagKey", k).Str("tagValue", *v).Msgf("Tags from %s nodegroup", *output.Nodegroup.NodegroupName)
	}

	return *output, nil
}

func HasNodegroupTag(tag string, nodegroupTags map[string]*string, ctx context.Context) (bool, error) {
	logWithContext := log.Ctx(ctx).With().Str("function", "HasNodegroupTag").Logger()

	for k, v := range nodegroupTags {
		tagTemp := k + ":" + *v
		if tag == tagTemp {
			logWithContext.Debug().Str("tagVar", tag).Str("nodegroupTag", tagTemp).Msg("Flag tag is found for nodegroup")

			return true, nil
		}
		logWithContext.Debug().Str("tagVar", tag).Str("nodegroupTag", tagTemp).Msg("Flag tag is not the same as checking nodegroup tag")
	}

	return false, nil
}
