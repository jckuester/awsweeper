package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AutoScalingGroup struct {
	Name string
	Tags map[string]string
}

func TestAWS_Resources_AutoScalingGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	autoscalingGroup := AutoScalingGroup{
		Name: "test-auto-scaling-group",
	}
	awsMock := createAutoScalingGroupAPIMock(mockCtrl, autoscalingGroup)

	// when
	resources, err := awsMock.RawResources(resource.AutoscalingGroup)
	require.NoError(t, err)
	groups := resources.([]*autoscaling.Group)

	// then
	assert.Len(t, groups, 1)
	assert.Equal(t, *groups[0].AutoScalingGroupName, autoscalingGroup.Name)
}

func createAutoScalingGroupAPIMock(mockCtrl *gomock.Controller, autoscalingGroup AutoScalingGroup) *resource.AWS {
	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)
	awsMock := &resource.AWS{
		AutoScalingAPI: mockObj,
	}

	var tags []*autoscaling.TagDescription
	for key, value := range autoscalingGroup.Tags {
		tags = append(tags, &autoscaling.TagDescription{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	mockObj.EXPECT().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{}).Return(
		&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{
				{
					AutoScalingGroupName: &autoscalingGroup.Name,
					Tags:                 tags,
				},
			},
		},
		nil)
	return awsMock
}
