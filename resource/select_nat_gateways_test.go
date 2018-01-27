package resource

import (
	"testing"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	natGateway = "aws_nat_gateway"
	yml2       = YamlCfg{
		natGateway: {
			Ids: []*string{aws.String("some-id")},
		},
	}

	f2 = &YamlFilter{
		cfg: yml2,
	}
)

func TestSelectNatGateways(t *testing.T) {
	c := &AWSClient{}

	ngwId := "some-id"
	ngws := &ec2.DescribeNatGatewaysOutput{
		NatGateways: []*ec2.NatGateway{
			{
				NatGatewayId: aws.String(ngwId),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(tagKey),
						Value: aws.String(tagValue),
					},
				},
				State: aws.String("available"),
			},
		},
	}

	res := Resources{
		{
			Type: "aws_nat_gateway",
			Id:   ngwId,
		},
	}

	resList := filterNatGateways(res, ngws, f2, c)

	for res := range resList {
		fmt.Println(res)
	}
}
