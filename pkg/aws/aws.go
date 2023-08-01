package aws

import (
	"fmt"

	awsLib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func EksClientSetup(awsRegion string) (*eks.EKS, error) {
	session, err := session.NewSession(&awsLib.Config{
		Region: awsLib.String(awsRegion)},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating new Eks session in %s region: %w", awsRegion, err)
	}

	return eks.New(session), nil
}

func SsmClientSetup(awsRegion string) (*ssm.SSM, error) {
	session, err := session.NewSession(&awsLib.Config{
		Region: awsLib.String(awsRegion)},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating new Ssm session in %s region: %w", awsRegion, err)
	}

	return ssm.New(session), nil
}

func Ec2ClientSetup() (*ec2.EC2, error) {
	session, err := session.NewSession(&awsLib.Config{})
	if err != nil {
		return nil, fmt.Errorf("error creating new Ec2 session: %w", err)
	}

	return ec2.New(session), nil
}
