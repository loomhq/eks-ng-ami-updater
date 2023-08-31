# EKS Node Group AMI Updater

[![Validation](https://github.com/loomhq/eks-ng-ami-updater/actions/workflows/validate.yml/badge.svg)](https://github.com/loomhq/eks-ng-ami-updater/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](https://opensource.org/license/mit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/loomhq/eks-ng-ami-updater)](https://goreportcard.com/badge/github.com/loomhq/eks-ng-ami-updater)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Floomhq%2Feks-ng-ami-updater.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Floomhq%2Feks-ng-ami-updater?ref=badge_shield)

EKS NG AMI Updater is an open source project that can be used to update kubernetes node group images. It is deployed as cronjob that runs weekly. By default it will find all node groups in all your EKS clusters and update them to the newest node group AMI if there is one available.

## Installation

To install the EKS NG AMI Updater in your cluster, execute the following commands:

#### Create IAM role

Create AWS role named `eks-nk-ami-updater` within below policy:

```
"Version": "2012-10-17",
"Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "ec2:DescribeRegions",
            "ec2:DescribeLaunchTemplateVersions",
            "ec2:RunInstances",
            "ec2:CreateTags"
        ],
        "Resource": "*"
    },
    {
        "Effect": "Allow",
        "Action": [
            "eks:DescribeNodegroup",
            "eks:ListNodegroups",
            "eks:ListClusters",
            "eks:UpdateNodegroupVersion"
        ],
        "Resource": "*"
    },
    {
        "Effect": "Allow",
        "Action": "ssm:GetParameter",
        "Resource": "*"
    }
]
```

#### Deploy application

Use below command to deploy EKS NG AMI Updater into your cluster (replace `<AWSACCOUNTID>` with your AWS account ID):

`helm install eks-nk-ami-updater oci://public.ecr.aws/loom/eks-ng-ami-updater --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::<AWSACCOUNTID>:role/eks-ng-ami-updater"`

## Parameters

| Command line flags     | Value keys                      | Type   | Default      | Description                                                                                                                                  |
| ---------------------- | ------------------------------- | ------ | ------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| --debug                | cmdOptions.debug                | bool   | false        | set log level to debug (eg. `--debug=true`)                                                                                                  |
| --dryrun               | cmdOptions.dryrun               | bool   | false        | set dryrun mode (eg. `--dryrun=true`)                                                                                                        |
| --nodegroups           | cmdOptions.nodegroups           | string | ""           | limit update amis to specified nodegroups (eg. `--nodegroups=eu-west-1:cluster-1:ngMain,eu-west-2:clusterStage:nodegroupStage1`)             |
| --regions              | cmdOptions.regions              | string | ""           | limit update amis to nodegroups from specified regions only (eg. `--regions=eu-west-1,us-west-1`)                                            |
| --skip-newer-than-days | cmdOptions.skip-newer-than-days | int    | 0            | skip ami update if the latest available in AWS ami image was published in less than provided number of days (eg. `--skip-newer-than-days=7`) |
| --tag                  | cmdOptions.tag                  | string | ""           | update amis only for nodegroups within this tag (eg. `--tag=env:production`)                                                                 |
| n/a                    | schedule                        | string | "30 7 * * 0" | schedule run within [cron syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax)                      |

All flags are connected with the AND operator. E.g. if we use two such flags `"--regions=eu-west-1 --nodegroups=us-west-2:cluster-1:nodegroup1"` then no images will be updated due to mismatched regions.

## Examples

`eks-ng-ami-updater --regions=us-west-1,us-west-2 --tag=env:production` - all nodes from any node groups from any clusters which are run in us-west-1 or us-west-1 region AND which have env tag set to production, will be updated.

`eks-ng-ami-updater --nodegroups=eu-west-1:cluster-1:ngMain --skip-newer-than-days=7` - all nodes from 'ngMain' node group from 'cluster-1' cluster will be updated but only if the newest availiable AMI is older than 7 days.

## FAQ

**Q:** I want to run updates in a testing environment first then in production a few days later. How do I do that? \
**A**: You can deploy EKS NG AMI Updater twice. The first one can be configured to do updates only for testing clusters and the second one can be configured (different 'schedule' definition) only for production.

**Q:** I don't want to be the first person in the world to use a new just-published AWS AMI image. What can I do? \
**A:** You can use the `skip-newer-than-days` parameter to define a delay (we recommend max 7 days delay).

**Q:** I set `skip-newer-than-days` parameter to 60 days and my AMI images haven't been updated for last 80 days. Is this normal? \
**A:** EKS NG AMI Updater is checking the release date for the last (newest) published by AWS AMI image. So if e.g. AWS releases new AMI images every 20 days than the newest availiable AWS AMI image will be always newer than the 60 day delay that you set. This is why we recommend setting the 'skip-newer-than-days' parameter to a max of 7 days.

## Maintainers

This project was created by [Andrzej Wisniewski](https://github.com/AndrzejWisniewski) at [Loom](https://github.com/loomhq/).


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Floomhq%2Feks-ng-ami-updater.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Floomhq%2Feks-ng-ami-updater?ref=badge_large)