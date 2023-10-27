package updater

import (
	"context"
	"strings"
	"time"

	"github.com/loomhq/eks-ng-ami-updater/pkg/aws"
	"github.com/loomhq/eks-ng-ami-updater/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func GetNodeGroupsToUpdateAmi(skipNewerThanDays uint, regionsVar, nodegroupsVar []string, tagVar string, ctx context.Context) ([]aws.NodeGroup, error) {
	var nodegroupsToUpdateAmi []aws.NodeGroup
	var nodegroupsFromRegion []aws.NodeGroup
	var nodegroupsReadyForAmiUpdate []aws.NodeGroup
	var regions []string
	var isOldEnough bool
	var nodegroupHasTag bool
	var regionIsAllowed bool

	logWithContext := log.Ctx(ctx).With().Str("function", "GetNodeGroupsToUpdateAmi").Logger()

	if len(nodegroupsVar) > 0 {
		for _, s := range nodegroupsVar {
			nodegroupSplit := strings.Split(s, ":")
			regionIsAllowed = true
			if len(regionsVar) > 0 {
				regionIsAllowed = utils.Contains(regionsVar, nodegroupSplit[0])
			}
			if regionIsAllowed {
				nodegroupsToUpdateAmi = append(nodegroupsToUpdateAmi, aws.NodeGroup{Region: nodegroupSplit[0], ClusterName: nodegroupSplit[1], NodegroupName: nodegroupSplit[2]})
				logWithContext.Debug().Str("region", nodegroupSplit[0]).Str("cluster", nodegroupSplit[2]).Str("nodegroup", nodegroupSplit[1]).Strs("nodegrpoupsVar", nodegroupsVar).Strs("regionsVar", regionsVar).Msg("add nodegroup to the ami upgrade nodegroups checking list")
			} else {
				logWithContext.Debug().Str("region", nodegroupSplit[0]).Str("cluster", nodegroupSplit[2]).Str("nodegroup", nodegroupSplit[1]).Strs("nodegrpoupsVar", nodegroupsVar).Strs("regionsVar", regionsVar).Msg("nodegroup is not in the regions defined by 'regions' flag")
			}
		}
	} else {
		svcEc2, err := aws.Ec2ClientSetup()
		if err != nil {
			return nil, err
		}
		awsEc2 := aws.RealEc2{Svc: svcEc2}

		regions, err = aws.GetRegionsToCheck(regionsVar, awsEc2, ctx)
		if err != nil {
			return nil, err
		}

		for _, region := range regions {
			nodegroupsFromRegion, err = aws.GetNodegroupsFromRegion(region, ctx)
			if err != nil {
				return nil, err
			}
			nodegroupsToUpdateAmi = append(nodegroupsToUpdateAmi, nodegroupsFromRegion...)
			for _, nodegroup := range nodegroupsFromRegion {
				logWithContext.Debug().Str("region", nodegroup.Region).Str("cluster", nodegroup.ClusterName).Str("nodegroup", nodegroup.NodegroupName).Strs("regionsVar", regionsVar).Msg("add nodegroup to the ami upgrade nodegroups checking list")
			}
		}
	}

	for _, nodegroup := range nodegroupsToUpdateAmi {
		svcEks, err := aws.EksClientSetup(nodegroup.Region)
		if err != nil {
			return nil, err
		}
		awsEks := aws.RealEks{Svc: svcEks}

		svcSsm, err := aws.SsmClientSetup(nodegroup.Region)
		if err != nil {
			return nil, err
		}
		awsSsm := aws.RealSsm{Svc: (svcSsm)}

		nodegroupDescription, err := aws.GetNodegroupDescription(nodegroup, awsEks, ctx)
		if err != nil {
			return nil, err
		}

		nodegroupHasTag = true
		if tagVar != "" {
			nodegroupHasTag, err = aws.HasNodegroupTag(tagVar, nodegroupDescription.Nodegroup.Tags, ctx)
			if err != nil {
				return nil, err
			}
			if !nodegroupHasTag {
				logWithContext.Debug().Str("region", nodegroup.Region).Str("cluster", nodegroup.ClusterName).Str("nodegroup", nodegroup.NodegroupName).Msgf("skip ami update for this nodegroup ('%s' tag is not exist)", tagVar)
			}
		}

		isOldEnough = true
		if skipNewerThanDays > 0 && nodegroupHasTag {
			today := time.Now()
			isOldEnough, err = aws.IsLastAmiOldEnough(skipNewerThanDays, nodegroup, today, *nodegroupDescription.Nodegroup.AmiType, *nodegroupDescription.Nodegroup.Version, awsSsm, ctx)
			if err != nil {
				return nil, err
			}
			if !isOldEnough {
				logWithContext.Debug().Str("region", nodegroup.Region).Str("cluster", nodegroup.ClusterName).Str("nodegroup", nodegroup.NodegroupName).Msg("skip ami update for this nodegroup (latest available ami for this nodegroup is too new)")
			}
		}

		if isOldEnough && nodegroupHasTag {
			nodegroupsReadyForAmiUpdate = append(nodegroupsReadyForAmiUpdate, nodegroup)
			logWithContext.Info().Str("region", nodegroup.Region).Str("cluster", nodegroup.ClusterName).Str("nodegroup", nodegroup.NodegroupName).Msg("nodegroup is ready for update")
		}
	}

	if len(nodegroupsReadyForAmiUpdate) == 0 {
		logWithContext.Info().Msg("no nodegroups are ready for ami update")
	}

	return nodegroupsReadyForAmiUpdate, nil
}

func UpdateAmi(dryrun bool, skipNewerThanDays uint, regionsVar, nodegroupsVar []string, tagVar string, ctx context.Context) error {
	var errorGroup errgroup.Group

	nodegroups, err := GetNodeGroupsToUpdateAmi(skipNewerThanDays, regionsVar, nodegroupsVar, tagVar, ctx)
	if err != nil {
		return err
	}

	for _, nodegroup := range nodegroups {
		nodegroup := nodegroup
		errorGroup.Go(func() error {
			return aws.AmiUpdate(nodegroup.Region, nodegroup.ClusterName, nodegroup.NodegroupName, dryrun, ctx)
		})
	}

	return errors.Wrap(errorGroup.Wait(), "at least one nodegroup can not be updated")
}
