package resource

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWS_Resources_AutoScalingGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)

	awsMock := &AWS{
		AutoScalingAPI: mockObj,
	}

	autoscalingGroupName := "bla"

	mockObj.EXPECT().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{}).Return(
		&autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []*autoscaling.Group{
				{
					AutoScalingGroupName: &autoscalingGroupName,
				},
			},
		},
		nil)

	// when
	resources, err := awsMock.rawResources(AutoscalingGroup)
	require.NoError(t, err)
	groups := resources.([]*autoscaling.Group)

	// then
	assert.Len(t, groups, 1)
	assert.Equal(t, *groups[0].AutoScalingGroupName, autoscalingGroupName)
}
