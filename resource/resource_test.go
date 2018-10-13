package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	someVpcID = "some-vpc-id"
	tagKey    = "bla"
	tagValue  = "blub"

	otherVpcID = "other-vpc-id"

	vpcs = []*ec2.Vpc{
		{
			VpcId: aws.String(someVpcID),
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(tagKey),
					Value: aws.String(tagValue),
				},
			},
		},
		{
			VpcId: aws.String(otherVpcID),
		},
	}
)

func TestAWS_DeletableResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	autoscalingGroup := AutoScalingGroup{
		Name: "test-auto-scaling-group",
		Tags: map[string]string{
			"test-tag-key": "test-tag-value",
		},
	}
	awsMock := createAutoScalingGroupAPIMock(mockCtrl, autoscalingGroup)

	rawResources, err := awsMock.RawResources(resource.AutoscalingGroup)
	require.NoError(t, err)

	// when
	res, err := awsMock.DeletableResources(resource.AutoscalingGroup, rawResources)
	require.NoError(t, err)

	require.Len(t, res, 1)
	require.Equal(t, res[0].ID, autoscalingGroup.Name)
	require.Equal(t, res[0].Tags, autoscalingGroup.Tags)
}
