package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ssm"
)

type SSM interface {
	GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

type RealSsm struct {
	Svc *ssm.SSM
}

func (t RealSsm) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	result, err := t.Svc.GetParameter(input)
	if err != nil {
		return nil, fmt.Errorf("error getting parameters: %w", err)
	}

	return result, nil
}
